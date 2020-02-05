package aria2

import (
	"context"
	model "github.com/HFO4/cloudreve/models"
	"github.com/zyxar/argo/rpc"
	"path/filepath"
	"strconv"
	"time"
)

// RPCService 通过RPC服务的Aria2任务管理器
type RPCService struct {
	options *clientOptions
	caller  rpc.Client
}

type clientOptions struct {
	Options []interface{} // 创建下载时额外添加的设置
}

// Init 初始化
func (client *RPCService) Init(server, secret string, timeout int, options []interface{}) error {
	// 客户端已存在，则关闭先前连接
	if client.caller != nil {
		client.caller.Close()
	}

	client.options = &clientOptions{
		Options: options,
	}
	caller, err := rpc.New(context.Background(), server, secret, time.Duration(timeout)*time.Second,
		EventNotifier)
	client.caller = caller
	return err
}

// Status 查询下载状态
func (client *RPCService) Status(task *model.Download) (rpc.StatusInfo, error) {
	return client.caller.TellStatus(task.GID)
}

// CreateTask 创建新任务
func (client *RPCService) CreateTask(task *model.Download) error {
	// 生成存储路径
	task.Path = filepath.Join(
		model.GetSettingByName("aria2_temp_path"),
		"aria2",
		strconv.FormatInt(time.Now().UnixNano(), 10),
	)

	// 创建下载任务
	options := []interface{}{map[string]string{"dir": task.Path}}
	if len(client.options.Options) > 0 {
		options = append(options, client.options.Options)
	}
	gid, err := client.caller.AddURI(task.Source, options...)
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
