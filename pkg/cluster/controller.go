package cluster

import (
	"bytes"
	"encoding/gob"
	"fmt"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/common"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/rpc"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/mq"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/jinzhu/gorm"
	"net/url"
	"sync"
)

var DefaultController Controller

// Controller controls communications between master and slave
type Controller interface {
	// Handle heartbeat sent from master
	HandleHeartBeat(*serializer.NodePingReq) (serializer.NodePingResp, error)

	// Get Aria2 Instance by master node ID
	GetAria2Instance(string) (common.Aria2, error)

	// Send event change message to master node
	SendNotification(string, string, mq.Message) error

	// Submit async task into task pool
	SubmitTask(string, interface{}, string, func(interface{})) error

	// Get master node info
	GetMasterInfo(string) (*MasterInfo, error)

	// Get master OneDrive policy credential
	GetOneDriveToken(string, uint) (string, error)
}

type slaveController struct {
	masters map[string]MasterInfo
	lock    sync.RWMutex
}

// info of master node
type MasterInfo struct {
	ID  string
	TTL int
	URL *url.URL
	// used to invoke aria2 rpc calls
	Instance Node
	Client   request.Client

	jobTracker map[string]bool
}

func InitController() {
	DefaultController = &slaveController{
		masters: make(map[string]MasterInfo),
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
			origin.Instance.Kill()
		}

		masterUrl, err := url.Parse(req.SiteURL)
		if err != nil {
			return serializer.NodePingResp{}, err
		}

		c.masters[req.SiteID] = MasterInfo{
			ID:  req.SiteID,
			URL: masterUrl,
			TTL: req.CredentialTTL,
			Client: request.NewClient(
				request.WithEndpoint(masterUrl.String()),
				request.WithSlaveMeta(fmt.Sprintf("%d", req.Node.ID)),
				request.WithCredential(auth.HMACAuth{
					SecretKey: []byte(req.Node.MasterKey),
				}, int64(req.CredentialTTL)),
			),
			jobTracker: make(map[string]bool),
			Instance: NewNodeFromDBModel(&model.Node{
				Model:                  gorm.Model{ID: req.Node.ID},
				MasterKey:              req.Node.MasterKey,
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
		return node.Instance.GetAria2Instance(), nil
	}

	return nil, ErrMasterNotFound
}

func (c *slaveController) SendNotification(id, subject string, msg mq.Message) error {
	c.lock.RLock()

	if node, ok := c.masters[id]; ok {
		c.lock.RUnlock()

		body := bytes.Buffer{}
		enc := gob.NewEncoder(&body)
		if err := enc.Encode(&msg); err != nil {
			return err
		}

		res, err := node.Client.Request(
			"PUT",
			fmt.Sprintf("/api/v3/slave/notification/%s", subject),
			&body,
		).CheckHTTPResponse(200).DecodeResponse()
		if err != nil {
			return err
		}

		if res.Code != 0 {
			return serializer.NewErrorFromResponse(res)
		}

		return nil
	}

	c.lock.RUnlock()
	return ErrMasterNotFound
}

// SubmitTask 提交异步任务
func (c *slaveController) SubmitTask(id string, job interface{}, hash string, submitter func(interface{})) error {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if node, ok := c.masters[id]; ok {
		if _, ok := node.jobTracker[hash]; ok {
			// 任务已存在，直接返回
			return nil
		}

		node.jobTracker[hash] = true
		submitter(job)
		return nil
	}

	return ErrMasterNotFound
}

// GetMasterInfo 获取主机节点信息
func (c *slaveController) GetMasterInfo(id string) (*MasterInfo, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if node, ok := c.masters[id]; ok {
		return &node, nil
	}

	return nil, ErrMasterNotFound
}

// GetOneDriveToken 获取主机OneDrive凭证
func (c *slaveController) GetOneDriveToken(id string, policyID uint) (string, error) {
	c.lock.RLock()

	if node, ok := c.masters[id]; ok {
		c.lock.RUnlock()

		res, err := node.Client.Request(
			"GET",
			fmt.Sprintf("/api/v3/slave/credential/onedrive/%d", policyID),
			nil,
		).CheckHTTPResponse(200).DecodeResponse()
		if err != nil {
			return "", err
		}

		if res.Code != 0 {
			return "", serializer.NewErrorFromResponse(res)
		}

		return res.Data.(string), nil
	}

	c.lock.RUnlock()
	return "", ErrMasterNotFound
}
