package cluster

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"sync"
)

type SlaveNode struct {
	Model    *model.Node
	callback func()
	lock     sync.RWMutex
}

// IsFeatureEnabled 查询节点的某项功能是否启用
func (node *SlaveNode) IsFeatureEnabled(feature string) bool {
	switch feature {
	default:
		return false
	}
}

// SubscribeStatusChange 订阅节点状态更改
func (node *SlaveNode) SubscribeStatusChange(callback func()) {
	node.lock.Lock()
	node.callback = callback
	node.lock.Unlock()
}

// PingLoop
// RecoverLoop
