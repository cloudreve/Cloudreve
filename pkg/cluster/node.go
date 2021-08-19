package cluster

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
)

type Node interface {
	Init(node *model.Node)
	IsFeatureEnabled(feature string) bool
	SubscribeStatusChange(callback func(isActive bool, id uint))
	Ping(req *serializer.NodePingReq) (*serializer.NodePingResp, error)
	IsActive() bool
	GetAria2Instance() (aria2.Aria2, error)
}

func getNodeFromDBModel(node *model.Node) Node {
	switch node.Type {
	case model.SlaveNodeType:
		slave := &SlaveNode{}
		slave.Init(node)
		return slave
	default:
		master := &MasterNode{}
		master.Init(node)
		return master
	}
}
