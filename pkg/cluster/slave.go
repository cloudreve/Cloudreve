package cluster

import (
	"encoding/json"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/common"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/rpc"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"io"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
)

type SlaveNode struct {
	Model        *model.Node
	AuthInstance auth.Auth
	Active       bool

	caller   slaveCaller
	callback func(bool, uint)
	close    chan bool
	lock     sync.RWMutex
}

type slaveCaller struct {
	parent *SlaveNode
	Client request.Client
}

// Init 初始化节点
func (node *SlaveNode) Init(nodeModel *model.Node) {
	node.lock.Lock()
	node.Model = nodeModel
	node.AuthInstance = auth.HMACAuth{SecretKey: []byte(nodeModel.SlaveKey)}
	node.caller.Client = request.HTTPClient{}
	node.caller.parent = node
	node.Active = true
	if node.close != nil {
		node.close <- true
	}
	node.lock.Unlock()

	go node.StartPingLoop()
}

// IsFeatureEnabled 查询节点的某项功能是否启用
func (node *SlaveNode) IsFeatureEnabled(feature string) bool {
	node.lock.RLock()
	defer node.lock.RUnlock()

	switch feature {
	case "aria2":
		return node.Model.Aria2Enabled
	default:
		return false
	}
}

// SubscribeStatusChange 订阅节点状态更改
func (node *SlaveNode) SubscribeStatusChange(callback func(bool, uint)) {
	node.lock.Lock()
	node.callback = callback
	node.lock.Unlock()
}

// Ping 从机节点，返回从机负载
func (node *SlaveNode) Ping(req *serializer.NodePingReq) (*serializer.NodePingResp, error) {
	reqBodyEncoded, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	bodyReader := strings.NewReader(string(reqBodyEncoded))
	signTTL := model.GetIntSetting("slave_api_timeout", 60)

	resp, err := node.caller.Client.Request(
		"POST",
		node.getAPIUrl("heartbeat"),
		bodyReader,
		request.WithMasterMeta(),
		request.WithTimeout(time.Duration(signTTL)*time.Second),
		request.WithCredential(node.AuthInstance, int64(signTTL)),
	).CheckHTTPResponse(200).DecodeResponse()
	if err != nil {
		return nil, err
	}

	// 处理列取结果
	if resp.Code != 0 {
		return nil, serializer.NewErrorFromResponse(resp)
	}

	var res serializer.NodePingResp

	if resStr, ok := resp.Data.(string); ok {
		err = json.Unmarshal([]byte(resStr), &res)
		if err != nil {
			return nil, err
		}
	}

	return &res, nil
}

// IsActive 返回节点是否在线
func (node *SlaveNode) IsActive() bool {
	node.lock.RLock()
	defer node.lock.RUnlock()

	return node.Active
}

// Kill 结束节点内相关循环
func (node *SlaveNode) Kill() {
	node.lock.RLock()
	defer node.lock.RUnlock()

	if node.close != nil {
		close(node.close)
	}
}

// GetAria2Instance 获取从机Aria2实例
func (node *SlaveNode) GetAria2Instance() common.Aria2 {
	node.lock.RLock()
	defer node.lock.RUnlock()

	if !node.Model.Aria2Enabled {
		return &common.DummyAria2{}
	}

	return &node.caller
}

func (node *SlaveNode) ID() uint {
	node.lock.RLock()
	defer node.lock.RUnlock()

	return node.Model.ID
}

// getAPIUrl 获取接口请求地址
func (node *SlaveNode) getAPIUrl(scope string) string {
	node.lock.RLock()
	serverURL, err := url.Parse(node.Model.Server)
	node.lock.RUnlock()
	if err != nil {
		return ""
	}

	var controller *url.URL
	controller, _ = url.Parse("/api/v3/slave")
	controller.Path = path.Join(controller.Path, scope)
	return serverURL.ResolveReference(controller).String()
}

func (node *SlaveNode) changeStatus(isActive bool) {
	node.lock.RLock()
	id := node.Model.ID
	if isActive != node.Active {
		node.lock.RUnlock()
		node.lock.Lock()
		node.Active = isActive
		node.lock.Unlock()
		node.callback(isActive, id)
	} else {
		node.lock.RUnlock()
	}

}

func (node *SlaveNode) StartPingLoop() {
	node.lock.Lock()
	node.close = make(chan bool)
	node.lock.Unlock()

	tickDuration := time.Duration(model.GetIntSetting("slave_ping_interval", 300)) * time.Second
	recoverDuration := time.Duration(model.GetIntSetting("slave_recover_interval", 600)) * time.Second
	pingTicker := time.Duration(0)

	util.Log().Debug("从机节点 [%s] 启动心跳循环", node.Model.Name)
	retry := 0
	recoverMode := false

loop:
	for {
		select {
		case <-time.After(pingTicker):
			if pingTicker == 0 {
				pingTicker = tickDuration
			}

			util.Log().Debug("从机节点 [%s] 发送Ping", node.Model.Name)
			res, err := node.Ping(node.getHeartbeatContent(false))
			if err != nil {
				util.Log().Debug("Ping从机节点 [%s] 时发生错误: %s", node.Model.Name, err)
				retry++
				if retry >= model.GetIntSetting("slave_node_retry", 3) {
					util.Log().Debug("从机节点 [%s] Ping 重试已达到最大限制，将从机节点标记为不可用", node.Model.Name)
					node.changeStatus(false)

					if !recoverMode {
						// 启动恢复监控循环
						util.Log().Debug("从机节点 [%s] 进入恢复模式", node.Model.Name)
						pingTicker = recoverDuration
						recoverMode = true
					}
				}
			} else {
				if recoverMode {
					util.Log().Debug("从机节点 [%s] 复活", node.Model.Name)
					pingTicker = tickDuration
					recoverMode = false
				}

				util.Log().Debug("从机节点 [%s] 状态: %s", node.Model.Name, res)
				node.changeStatus(true)
				retry = 0
			}

		case <-node.close:
			util.Log().Debug("从机节点 [%s] 收到关闭信号", node.Model.Name)
			break loop
		}
	}
}

// getHeartbeatContent gets serializer.NodePingReq used to send heartbeat to slave
func (node *SlaveNode) getHeartbeatContent(isUpdate bool) *serializer.NodePingReq {
	return &serializer.NodePingReq{
		IsUpdate: isUpdate,
		SiteID:   model.GetSettingByName("siteID"),
		Node:     node.Model,
	}
}

func (s *slaveCaller) Init() error {
	return nil
}

// SendAria2Call send remote aria2 call to slave node
func (s *slaveCaller) SendAria2Call(body *serializer.SlaveAria2Call, scope string) (*serializer.Response, error) {
	reqReader, err := getAria2RequestBody(body)
	if err != nil {
		return nil, err
	}

	signTTL := model.GetIntSetting("slave_api_timeout", 60)
	return s.Client.Request(
		"POST",
		s.parent.getAPIUrl("aria2/"+scope),
		reqReader,
		request.WithMasterMeta(),
		request.WithTimeout(time.Duration(signTTL)*time.Second),
		request.WithCredential(s.parent.AuthInstance, int64(signTTL)),
	).CheckHTTPResponse(200).DecodeResponse()
}

func (s *slaveCaller) CreateTask(task *model.Download, options map[string]interface{}) (string, error) {
	s.parent.lock.RLock()
	defer s.parent.lock.RUnlock()

	req := &serializer.SlaveAria2Call{
		Task:         task,
		GroupOptions: options,
	}

	res, err := s.SendAria2Call(req, "task")
	if err != nil {
		return "", err
	}

	if res.Code != 0 {
		return "", serializer.NewErrorFromResponse(res)
	}

	return res.Data.(string), err
}

func (s *slaveCaller) Status(task *model.Download) (rpc.StatusInfo, error) {
	panic("implement me")
}

func (s *slaveCaller) Cancel(task *model.Download) error {
	panic("implement me")
}

func (s *slaveCaller) Select(task *model.Download, files []int) error {
	panic("implement me")
}

func (s *slaveCaller) GetConfig() model.Aria2Option {
	s.parent.lock.RLock()
	defer s.parent.lock.RUnlock()

	return s.parent.Model.Aria2OptionsSerialized
}

func getAria2RequestBody(body *serializer.SlaveAria2Call) (io.Reader, error) {
	reqBodyEncoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return strings.NewReader(string(reqBodyEncoded)), nil
}
