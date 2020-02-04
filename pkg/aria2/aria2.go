package aria2

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/pkg/util"
	"net/url"
)

// Instance 默认使用的Aria2处理实例
var Instance Aria2 = &DummyAria2{}

// Aria2 离线下载处理接口
type Aria2 interface {
	// CreateTask 创建新的任务
	CreateTask(task *model.Download) error
}

const (
	// URLTask 从URL添加的任务
	URLTask = iota
	// TorrentTask 种子任务
	TorrentTask
)

const (
	// Ready 准备就绪
	Ready = iota
)

var (
	// ErrNotEnabled 功能未开启错误
	ErrNotEnabled = serializer.NewError(serializer.CodeNoPermissionErr, "离线下载功能未开启", nil)
)

// DummyAria2 未开启Aria2功能时使用的默认处理器
type DummyAria2 struct {
}

// CreateTask 创建新任务，此处直接返回未开启错误
func (instance *DummyAria2) CreateTask(task *model.Download) error {
	return ErrNotEnabled
}

// Init 初始化
func Init() {
	options := model.GetSettingByNames("aria2_rpcurl", "aria2_token", "aria2_options")
	timeout := model.GetIntSetting("aria2_call_timeout", 5)
	if options["aria2_rpcurl"] == "" {
		// 未开启Aria2服务
		return
	}

	util.Log().Info("初始化 aria2 RPC 服务[%s]", options["aria2_rpcurl"])
	client := &RPCService{}
	if previousClient, ok := Instance.(*RPCService); ok {
		client = previousClient
	}

	// 解析RPC服务地址
	server, err := url.Parse(options["aria2_rpcurl"])
	if err != nil {
		util.Log().Warning("无法解析 aria2 RPC 服务地址，%s", err)
		return
	}
	server.Path = "/jsonrpc"

	// todo 加载自定义下载配置
	if err := client.Init(server.String(), options["aria2_token"], timeout, []interface{}{}); err != nil {
		util.Log().Warning("初始化 aria2 RPC 服务失败，%s", err)
		return
	}

	Instance = client
}
