package aria2

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2"
	"github.com/cloudreve/Cloudreve/v3/pkg/cluster"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/gin-gonic/gin"
)

// AddURLService 添加URL离线下载服务
type AddURLService struct {
	URL string `json:"url" binding:"required"`
	Dst string `json:"dst" binding:"required,min=1"`
}

// Add 创建新的链接离线下载任务
func (service *AddURLService) Add(c *gin.Context, taskType int) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}
	defer fs.Recycle()

	// 检查用户组权限
	if !fs.User.Group.OptionsSerialized.Aria2 {
		return serializer.Err(serializer.CodeGroupNotAllowed, "当前用户组无法进行此操作", nil)
	}

	// 存放目录是否存在
	if exist, _ := fs.IsPathExist(service.Dst); !exist {
		return serializer.Err(serializer.CodeNotFound, "存放路径不存在", nil)
	}

	// 创建任务
	task := &model.Download{
		Status: aria2.Ready,
		Type:   taskType,
		Dst:    service.Dst,
		UserID: fs.User.ID,
		Source: service.URL,
	}

	// 获取 Aria2 负载均衡器
	aria2.Lock.RLock()
	lb := aria2.LB
	aria2.Lock.RUnlock()

	// 获取 Aria2 实例
	err, node := cluster.Default.BalanceNodeByFeature("aria2", lb)
	if err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "Aria2 实例获取失败", err)
	}

	// 创建任务
	gid, err := node.GetAria2Instance().CreateTask(task, fs.User.Group.OptionsSerialized.Aria2Options)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, "任务创建失败", err)
	}

	task.GID = gid
	_, err = task.Create()
	if err != nil {
		return serializer.DBErr("任务创建失败", err)
	}

	// 创建任务监控
	aria2.NewMonitor(task)

	aria2.Lock.RUnlock()
	return serializer.Response{}
}
