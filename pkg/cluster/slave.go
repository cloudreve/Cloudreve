package cluster

import (
	"encoding/json"
	"errors"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/common"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
)

type SlaveNode struct {
	Model        *model.Node
	AuthInstance auth.Auth
	Client       request.Client
	Active       bool

	callback func(bool, uint)
	close    chan bool
	lock     sync.RWMutex
}

// Init 初始化节点
func (node *SlaveNode) Init(nodeModel *model.Node) {
	node.lock.Lock()
	node.Model = nodeModel
	node.AuthInstance = auth.HMACAuth{SecretKey: []byte(nodeModel.SecretKey)}
	node.Client = request.HTTPClient{}
	node.Active = true
	if node.close != nil {
		node.close <- true
	}
	node.lock.Unlock()

	go node.StartPingLoop()
}

// IsFeatureEnabled 查询节点的某项功能是否启用
func (node *SlaveNode) IsFeatureEnabled(feature string) bool {
	switch feature {
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

	resp, err := node.Client.Request(
		"POST",
		node.getAPIUrl("heartbeat"),
		bodyReader,
		request.WithCredential(node.AuthInstance, int64(signTTL)),
	).CheckHTTPResponse(200).DecodeResponse()
	if err != nil {
		return nil, err
	}

	// 处理列取结果
	if resp.Code != 0 {
		return nil, errors.New(resp.Error)
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
	return node.Active
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
	pingTicker := time.NewTicker(tickDuration)
	defer pingTicker.Stop()

	util.Log().Debug("从机节点 [%s] 启动心跳循环", node.Model.Name)
	retry := 0
	recoverMode := false

loop:
	for {
		select {
		case <-pingTicker.C:
			util.Log().Debug("从机节点 [%s] 发送Ping", node.Model.Name)
			res, err := node.Ping(&serializer.NodePingReq{})
			if err != nil {
				util.Log().Debug("Ping从机节点 [%s] 时发生错误: %s", node.Model.Name, err)
				retry++
				if retry >= model.GetIntSetting("slave_node_retry", 3) {
					util.Log().Debug("从机节点 [%s] Ping 重试已达到最大限制，将从机节点标记为不可用", node.Model.Name)
					node.changeStatus(false)

					if !recoverMode {
						// 启动恢复监控循环
						util.Log().Debug("从机节点 [%s] 进入恢复模式", node.Model.Name)
						pingTicker.Stop()
						pingTicker = time.NewTicker(recoverDuration)
						recoverMode = true
					}
				}
			} else {
				if recoverMode {
					util.Log().Debug("从机节点 [%s] 复活", node.Model.Name)
					pingTicker.Stop()
					pingTicker = time.NewTicker(tickDuration)
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

// GetAria2Instance 获取从机Aria2实例
func (node *SlaveNode) GetAria2Instance() common.Aria2 {
	return nil
}

func (node *SlaveNode) ID() uint {
	node.lock.RLock()
	defer node.lock.RUnlock()

	return node.Model.ID
}
