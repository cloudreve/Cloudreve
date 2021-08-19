package cluster

import (
	"encoding/json"
	"errors"
	model "github.com/cloudreve/Cloudreve/v3/models"
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

	active   bool
	callback func(bool, uint)
	close    chan bool
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
	return node.active
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
	if isActive != node.active {
		node.lock.RUnlock()
		node.lock.Lock()
		node.active = isActive
		node.lock.Unlock()
		node.callback(isActive, id)
	} else {
		node.lock.RUnlock()
	}

}

func (node *SlaveNode) StartPingLoop() {
	node.lock.Lock()
	node.close = make(chan bool)
	node.active = true
	node.lock.Unlock()

	t := time.NewTicker(time.Duration(model.GetIntSetting("slave_ping_interval", 300)) * time.Second)
	defer t.Stop()

	util.Log().Debug("从机节点 [%s] 启动心跳循环", node.Model.Name)
	retry := 0

loop:
	for {
		select {
		case <-t.C:
			util.Log().Debug("从机节点 [%s] 发送Ping", node.Model.Name)
			res, err := node.Ping(&serializer.NodePingReq{})
			if err != nil {
				util.Log().Debug("Ping从机节点 [%s] 时发生错误: %s", node.Model.Name, err)
				retry++
				if retry > model.GetIntSetting("slave_node_retry", 3) {
					util.Log().Debug("从机节点 [%s] Ping 重试已达到最大限制，将从机节点标记为不可用", node.Model.Name)
					node.changeStatus(false)
					break loop
				}
				break
			}

			util.Log().Debug("从机节点 [%s] 状态: %s", node.Model.Name, res)
			node.changeStatus(true)
		case <-node.close:
			util.Log().Debug("从机节点 [%s] 收到关闭信号", node.Model.Name)
			break loop
		}
	}
}
