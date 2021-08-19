package cluster

import (
	"context"
	"encoding/json"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/rpc"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"net/url"
	"sync"
	"time"
)

type MasterNode struct {
	Model    *model.Node
	aria2RPC rpcService
	lock     sync.RWMutex
}

// RPCService 通过RPC服务的Aria2任务管理器
type rpcService struct {
	Caller      rpc.Client
	Initialized bool

	parent  *MasterNode
	options *clientOptions
}

type clientOptions struct {
	Options map[string]interface{} // 创建下载时额外添加的设置
}

// Init 初始化节点
func (node *MasterNode) Init(nodeModel *model.Node) {
	node.lock.Lock()
	node.Model = nodeModel
	node.aria2RPC.parent = node
	node.lock.Unlock()

	node.lock.RLock()
	if node.Model.Aria2Enabled {
		node.lock.RUnlock()
		node.aria2RPC.Init()
		return
	}
	node.lock.RUnlock()
}

func (node *MasterNode) Ping(req *serializer.NodePingReq) (*serializer.NodePingResp, error) {
	return &serializer.NodePingResp{}, nil
}

// IsFeatureEnabled 查询节点的某项功能是否启用
func (node *MasterNode) IsFeatureEnabled(feature string) bool {
	switch feature {
	default:
		return false
	}
}

// SubscribeStatusChange 订阅节点状态更改
func (node *MasterNode) SubscribeStatusChange(callback func(isActive bool, id uint)) {
}

// IsActive 返回节点是否在线
func (node *MasterNode) IsActive() bool {
	return true
}

// GetAria2Instance 获取主机Aria2实例
func (node *MasterNode) GetAria2Instance() (aria2.Aria2, error) {
	if !node.Model.Aria2Enabled {
		return &aria2.DummyAria2{}, nil
	}

	node.lock.RLock()
	defer node.lock.RUnlock()
	if !node.aria2RPC.Initialized {
		return &aria2.DummyAria2{}, nil
	}

	return &node.aria2RPC, nil
}

func (r *rpcService) Init() error {
	r.parent.lock.Lock()
	defer r.parent.lock.Unlock()
	r.Initialized = false

	// 客户端已存在，则关闭先前连接
	if r.Caller != nil {
		r.Caller.Close()
	}

	// 解析RPC服务地址
	server, err := url.Parse(r.parent.Model.Aria2OptionsSerialized.Server)
	if err != nil {
		util.Log().Warning("无法解析主机 Aria2 RPC 服务地址，%s", err)
		return err
	}
	server.Path = "/jsonrpc"

	// 加载自定义下载配置
	var globalOptions map[string]interface{}
	err = json.Unmarshal([]byte(r.parent.Model.Aria2OptionsSerialized.Options), &globalOptions)
	if err != nil {
		util.Log().Warning("无法解析主机 Aria2 配置，%s", err)
		return err
	}

	r.options = &clientOptions{
		Options: globalOptions,
	}
	timeout := model.GetIntSetting("aria2_call_timeout", 5)
	caller, err := rpc.New(context.Background(), server.String(), r.parent.Model.Aria2OptionsSerialized.Token, time.Duration(timeout)*time.Second, aria2.EventNotifier)

	r.Caller = caller
	r.Initialized = true
	return err
}

func (r *rpcService) CreateTask(task *model.Download, options map[string]interface{}) (string, error) {
	panic("implement me")
}

func (r *rpcService) Status(task *model.Download) (rpc.StatusInfo, error) {
	panic("implement me")
}

func (r *rpcService) Cancel(task *model.Download) error {
	panic("implement me")
}

func (r *rpcService) Select(task *model.Download, files []int) error {
	panic("implement me")
}
