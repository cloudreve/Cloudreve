package aria2

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/common"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/monitor"
	"github.com/cloudreve/Cloudreve/v3/pkg/cluster"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/mq"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-gonic/gin"
)

// AddURLService 添加URL离线下载服务
type BatchAddURLService struct {
	URLs []string `json:"url" binding:"required"`
	Dst  string   `json:"dst" binding:"required,min=1"`
}

// Add 主机批量创建新的链接离线下载任务
func (service *BatchAddURLService) Add(c *gin.Context, taskType int) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	// 检查用户组权限
	if !fs.User.Group.OptionsSerialized.Aria2 {
		return serializer.Err(serializer.CodeGroupNotAllowed, "", nil)
	}

	// 存放目录是否存在
	if exist, _ := fs.IsPathExist(service.Dst); !exist {
		return serializer.Err(serializer.CodeParentNotExist, "", nil)
	}

	// 检查批量任务数量
	limit := fs.User.Group.OptionsSerialized.Aria2BatchSize
	if limit > 0 && len(service.URLs) > limit {
		return serializer.Err(serializer.CodeBatchAria2Size, "", nil)
	}

	res := make([]serializer.Response, 0, len(service.URLs))
	for _, target := range service.URLs {
		subService := &AddURLService{
			URL: target,
			Dst: service.Dst,
		}

		addRes := subService.Add(c, fs, taskType)
		res = append(res, addRes)
	}

	return serializer.Response{Data: res}
}

// AddURLService 添加URL离线下载服务
type AddURLService struct {
	URL           string `json:"url" binding:"required"`
	Dst           string `json:"dst" binding:"required,min=1"`
	PreferredNode uint   `json:"preferred_node"`
}

// Add 主机创建新的链接离线下载任务
func (service *AddURLService) Add(c *gin.Context, fs *filesystem.FileSystem, taskType int) serializer.Response {
	if fs == nil {
		var err error
		// 创建文件系统
		fs, err = filesystem.NewFileSystemFromContext(c)
		if err != nil {
			return serializer.Err(serializer.CodeCreateFSError, "", err)
		}
		defer fs.Recycle()

		// 检查用户组权限
		if !fs.User.Group.OptionsSerialized.Aria2 {
			return serializer.Err(serializer.CodeGroupNotAllowed, "", nil)
		}

		// 存放目录是否存在
		if exist, _ := fs.IsPathExist(service.Dst); !exist {
			return serializer.Err(serializer.CodeParentNotExist, "", nil)
		}
	}

	downloads := model.GetDownloadsByStatusAndUser(0, fs.User.ID, common.Downloading, common.Paused, common.Ready)
	limit := fs.User.Group.OptionsSerialized.Aria2BatchSize
	if limit > 0 && len(downloads)+1 > limit {
		return serializer.Err(serializer.CodeBatchAria2Size, "", nil)
	}

	if service.PreferredNode > 0 && !fs.User.Group.OptionsSerialized.SelectNode {
		return serializer.Err(serializer.CodeGroupNotAllowed, "not allowed to select nodes", nil)
	}

	// 创建任务
	task := &model.Download{
		Status: common.Ready,
		Type:   taskType,
		Dst:    service.Dst,
		UserID: fs.User.ID,
		Source: service.URL,
	}

	// 获取 Aria2 负载均衡器
	lb := aria2.GetLoadBalancer()

	// 获取 Aria2 实例
	err, node := cluster.Default.BalanceNodeByFeature("aria2", lb, fs.User.Group.OptionsSerialized.AvailableNodes,
		service.PreferredNode)
	if err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "Failed to get Aria2 instance", err)
	}

	// 创建任务
	gid, err := node.GetAria2Instance().CreateTask(task, fs.User.Group.OptionsSerialized.Aria2Options)
	if err != nil {
		return serializer.Err(serializer.CodeCreateTaskError, "", err)
	}

	task.GID = gid
	task.NodeID = node.ID()
	_, err = task.Create()
	if err != nil {
		return serializer.DBErr("Failed to create task record", err)
	}

	// 创建任务监控
	monitor.NewMonitor(task, cluster.Default, mq.GlobalMQ)

	return serializer.Response{}
}

// Add 从机创建新的链接离线下载任务
func Add(c *gin.Context, service *serializer.SlaveAria2Call) serializer.Response {
	caller, _ := c.Get("MasterAria2Instance")

	// 创建任务
	gid, err := caller.(common.Aria2).CreateTask(service.Task, service.GroupOptions)
	if err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "Failed to create aria2 task", err)
	}

	// 创建事件通知回调
	siteID, _ := c.Get("MasterSiteID")
	mq.GlobalMQ.SubscribeCallback(gid, func(message mq.Message) {
		if err := cluster.DefaultController.SendNotification(siteID.(string), message.TriggeredBy, message); err != nil {
			util.Log().Warning("Failed to send remote download task status change notifications: %s", err)
		}
	})

	return serializer.Response{Data: gid}
}
