package cluster

import (
	"context"
	"encoding/json"
	"errors"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
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

	callback func()
	ctx      context.Context
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

func (node *SlaveNode) pingLoop() {
	t := time.NewTicker(time.Second)

	for {
		select {
		case <-t.C:
		case <-node.ctx.Done():
		}
	}
}

// PingLoop
// RecoverLoop
