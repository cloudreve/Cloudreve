package slave

import (
	"encoding/gob"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/common"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/rpc"
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

	// Get Aria2 instance by master node id
	GetAria2Instance(string) (common.Aria2, error)
}

type slaveController struct {
	masters map[string]masterInfo
	lock    sync.RWMutex
}

// info of master node
type masterInfo struct {
	slaveID    uint
	id         string
	authClient auth.Auth
	// used to invoke aria2 rpc calls
	instance cluster.Node
}

func Init() {
	DefaultController = &slaveController{
		masters: make(map[string]masterInfo),
	}
	gob.Register(rpc.StatusInfo{})
}

func (c *slaveController) HandleHeartBeat(req *serializer.NodePingReq) (serializer.NodePingResp, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	req.Node.AfterFind()

	// close old node if exist
	origin, ok := c.masters[req.SiteID]

	if (ok && req.IsUpdate) || !ok {
		if ok {
			origin.instance.Kill()
		}

		c.masters[req.SiteID] = masterInfo{
			slaveID: req.Node.ID,
			id:      req.SiteID,
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

func (c *slaveController) GetAria2Instance(id string) (common.Aria2, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if node, ok := c.masters[id]; ok {
		return node.instance.GetAria2Instance(), nil
	}

	return nil, ErrMasterNotFound
}
