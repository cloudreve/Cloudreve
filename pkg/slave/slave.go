package slave

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/cluster"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"sync"
)

var DefaultController Controller

// Controller controls communications between master and slave
type Controller interface {
	// Handle heartbeat sent from master
	HandleHeartBeat(*serializer.NodePingReq) (serializer.NodePingResp, error)
}

type slaveController struct {
	masters map[string]masterInfo
	lock    sync.RWMutex
}

// info of master node
type masterInfo struct {
	slaveID    uint
	url        string
	authClient auth.Auth
	// used to invoke aria2 rpc calls
	instance cluster.Node
}

func Init() {
	DefaultController = &slaveController{
		masters: make(map[string]masterInfo),
	}
}

func (c *slaveController) HandleHeartBeat(req *serializer.NodePingReq) (serializer.NodePingResp, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	req.Node.AfterFind()

	// close old node if exist
	origin, ok := c.masters[req.MasterURL]

	if (ok && req.IsUpdate) || !ok {
		if ok {
			origin.instance.Kill()
		}

		c.masters[req.MasterURL] = masterInfo{
			slaveID: req.Node.ID,
			url:     req.MasterURL,
			authClient: auth.HMACAuth{
				SecretKey: []byte(req.Node.MasterKey),
			},
			instance: cluster.NewNodeFromDBModel(&model.Node{
				Type:                   model.MasterNodeType,
				Aria2Enabled:           req.Node.Aria2Enabled,
				Aria2OptionsSerialized: req.Node.Aria2OptionsSerialized,
			}),
		}
	}

	return serializer.NodePingResp{}, nil
}
