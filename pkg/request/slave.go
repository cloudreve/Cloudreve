package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"

	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
)

// TODO: move to slave pkg
// RemoteCallback 发送远程存储策略上传回调请求
func RemoteCallback(url string, body serializer.UploadCallback) error {
	callbackBody, err := json.Marshal(struct {
		Data serializer.UploadCallback `json:"data"`
	}{
		Data: body,
	})
	if err != nil {
		return serializer.NewError(serializer.CodeCallbackError, "无法编码回调正文", err)
	}

	resp := GeneralClient.Request(
		"POST",
		url,
		bytes.NewReader(callbackBody),
		WithTimeout(time.Duration(conf.SlaveConfig.CallbackTimeout)*time.Second),
		WithCredential(auth.General, int64(conf.SlaveConfig.SignatureTTL)),
	)

	if resp.Err != nil {
		return serializer.NewError(serializer.CodeCallbackError, "无法发起回调请求", resp.Err)
	}

	// 解析回调服务端响应
	resp = resp.CheckHTTPResponse(200)
	if resp.Err != nil {
		return serializer.NewError(serializer.CodeCallbackError, "服务器返回异常响应", resp.Err)
	}
	response, err := resp.DecodeResponse()
	if err != nil {
		return serializer.NewError(serializer.CodeCallbackError, "无法解析服务端返回的响应", err)
	}
	if response.Code != 0 {
		return serializer.NewError(response.Code, response.Msg, errors.New(response.Error))
	}

	return nil
}
