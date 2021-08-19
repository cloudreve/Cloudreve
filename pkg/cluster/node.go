package cluster

import model "github.com/cloudreve/Cloudreve/v3/models"

type Node interface {
	IsFeatureEnabled(feature string) bool
	SubscribeStatusChange(callback func())
}

func getNodeFromDBModel(node *model.Node) Node {
	switch node.Type {
	case model.SlaveNodeType:
		return &SlaveNode{
			Model: node,
		}
	default:
		return &MasterNode{
			Model: node,
		}
	}
}
