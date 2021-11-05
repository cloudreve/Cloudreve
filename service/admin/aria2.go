package admin

import (
	"bytes"
	"encoding/json"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"net/url"
	"time"

	"github.com/cloudreve/Cloudreve/v3/pkg/aria2"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
)

// Aria2TestService aria2连接测试服务
type Aria2TestService struct {
	Server string          `json:"server"`
	RPC    string          `json:"rpc" binding:"required"`
	Secret string          `json:"secret"`
	Token  string          `json:"token"`
	Type   model.ModelType `json:"type"`
}

// Test 测试aria2连接
func (service *Aria2TestService) TestMaster() serializer.Response {
	res, err := aria2.TestRPCConnection(service.RPC, service.Token, 5)
	if err != nil {
		return serializer.ParamErr(err.Error(), err)
	}

	if res.Version == "" {
		return serializer.ParamErr("RPC 服务返回非预期响应", nil)
	}

	return serializer.Response{Data: res.Version}
}

func (service *Aria2TestService) TestSlave() serializer.Response {
	slave, err := url.Parse(service.Server)
	if err != nil {
		return serializer.ParamErr("无法解析从机端地址，"+err.Error(), nil)
	}

	controller, _ := url.Parse("/api/v3/slave/ping/aria2")

	// 请求正文
	service.Type = model.MasterNodeType
	bodyByte, _ := json.Marshal(service)

	r := request.NewClient()
	res, err := r.Request(
		"POST",
		slave.ResolveReference(controller).String(),
		bytes.NewReader(bodyByte),
		request.WithTimeout(time.Duration(10)*time.Second),
		request.WithCredential(
			auth.HMACAuth{SecretKey: []byte(service.Secret)},
			int64(model.GetIntSetting("slave_api_timeout", 60)),
		),
	).DecodeResponse()
	if err != nil {
		return serializer.ParamErr("无连接到从机，"+err.Error(), nil)
	}

	if res.Code != 0 {
		return serializer.ParamErr("成功接到从机，但是从机返回："+res.Msg, nil)
	}

	return serializer.Response{Data: res.Data.(string)}
}
