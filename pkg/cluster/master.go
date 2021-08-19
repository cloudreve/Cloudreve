package cluster

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
)

type MasterNode struct {
	Model *model.Node
}

// IsFeatureEnabled 查询节点的某项功能是否启用
func (node *MasterNode) IsFeatureEnabled(feature string) bool {
	switch feature {
	default:
		return false
	}
}

// SubscribeStatusChange 订阅节点状态更改
func (node *MasterNode) SubscribeStatusChange(callback func()) {
}
