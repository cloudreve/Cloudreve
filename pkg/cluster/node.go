package cluster

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
)

type Node interface {
	IsFeatureEnabled(feature string) bool
	SubscribeStatusChange(callback func())
	Ping(req *serializer.NodePingReq) (*serializer.NodePingResp, error)
}

func getNodeFromDBModel(node *model.Node) Node {
	switch node.Type {
	case model.SlaveNodeType:
		return &SlaveNode{
			Model:        node,
			AuthInstance: auth.HMACAuth{SecretKey: []byte(node.SecretKey)},
			Client:       request.HTTPClient{},
		}
	default:
		return &MasterNode{
			Model: node,
		}
	}
}
