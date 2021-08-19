package cluster

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
)

type MasterNode struct {
	Model *model.Node
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

// InitAria2RPCClient 初始化主机 Aria2 RPC 服务
func (node *MasterNode) InitAria2RPCClient() error {
	return nil
}
