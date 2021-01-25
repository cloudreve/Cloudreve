package admin

import (
	"net/url"

	"github.com/cloudreve/Cloudreve/v3/pkg/aria2"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
)

// Aria2TestService aria2连接测试服务
type Aria2TestService struct {
	Server string `json:"server" binding:"required"`
	Token  string `json:"token"`
}

// Test 测试aria2连接
func (service *Aria2TestService) Test() serializer.Response {
	testRPC := aria2.RPCService{}

	// 解析RPC服务地址
	server, err := url.Parse(service.Server)
	if err != nil {
		return serializer.ParamErr("无法解析 aria2 RPC 服务地址, "+err.Error(), nil)
	}
	server.Path = "/jsonrpc"

	if err := testRPC.Init(server.String(), service.Token, 5, map[string]interface{}{}); err != nil {
		return serializer.ParamErr("无法初始化连接, "+err.Error(), nil)
	}

	defer testRPC.Caller.Close()

	info, err := testRPC.Caller.GetVersion()
	if err != nil {
		return serializer.ParamErr("无法请求 RPC 服务, "+err.Error(), nil)
	}

	if info.Version == "" {
		return serializer.ParamErr("RPC 服务返回非预期响应", nil)
	}

	return serializer.Response{Data: info.Version}
}
