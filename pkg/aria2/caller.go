package aria2

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/rpc"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

// RPCService 通过RPC服务的Aria2任务管理器
type RPCService struct {
	options *clientOptions
	Caller  rpc.Client
}

type clientOptions struct {
	Options map[string]interface{} // 创建下载时额外添加的设置
}

// Init 初始化
func (client *RPCService) Init(server, secret string, timeout int, options map[string]interface{}) error {
	// 客户端已存在，则关闭先前连接
	if client.Caller != nil {
		client.Caller.Close()
	}

	client.options = &clientOptions{
		Options: options,
	}
	caller, err := rpc.New(context.Background(), server, secret, time.Duration(timeout)*time.Second,
		EventNotifier)
	client.Caller = caller
	return err
}

// Status 查询下载状态
func (client *RPCService) Status(task *model.Download) (rpc.StatusInfo, error) {
	res, err := client.Caller.TellStatus(task.GID)
	if err != nil {
		// 失败后重试
		util.Log().Debug("无法获取离线下载状态，%s，10秒钟后重试", err)
		time.Sleep(time.Duration(10) * time.Second)
		res, err = client.Caller.TellStatus(task.GID)
	}

	return res, err
}

// Cancel 取消下载
func (client *RPCService) Cancel(task *model.Download) error {
	// 取消下载任务
	_, err := client.Caller.Remove(task.GID)
	if err != nil {
		util.Log().Warning("无法取消离线下载任务[%s], %s", task.GID, err)
	}

	//// 删除临时文件
	//util.Log().Debug("离线下载任务[%s]已取消，1 分钟后删除临时文件", task.GID)
	//go func(task *model.Download) {
	//	select {
	//	case <-time.After(time.Duration(60) * time.Second):
	//		err := os.RemoveAll(task.Parent)
	//		if err != nil {
	//			util.Log().Warning("无法删除离线下载临时目录[%s], %s", task.Parent, err)
	//		}
	//	}
	//}(task)

	return err
}

// Select 选取要下载的文件
func (client *RPCService) Select(task *model.Download, files []int) error {
	var selected = make([]string, len(files))
	for i := 0; i < len(files); i++ {
		selected[i] = strconv.Itoa(files[i])
	}
	_, err := client.Caller.ChangeOption(task.GID, map[string]interface{}{"select-file": strings.Join(selected, ",")})
	return err
}

// CreateTask 创建新任务
func (client *RPCService) CreateTask(task *model.Download, groupOptions map[string]interface{}) error {
	// 生成存储路径
	path := filepath.Join(
		model.GetSettingByName("aria2_temp_path"),
		"aria2",
		strconv.FormatInt(time.Now().UnixNano(), 10),
	)

	// 创建下载任务
	options := map[string]interface{}{
		"dir": path,
	}
	for k, v := range client.options.Options {
		options[k] = v
	}
	for k, v := range groupOptions {
		options[k] = v
	}

	gid, err := client.Caller.AddURI(task.Source, options)
	if err != nil || gid == "" {
		return err
	}

	// 保存到数据库
	task.GID = gid
	_, err = task.Create()
	if err != nil {
		return err
	}

	// 创建任务监控
	NewMonitor(task)

	return nil
}
