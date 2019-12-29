package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/HFO4/cloudreve/pkg/auth"
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/pkg/util"
	"io/ioutil"
	"time"
)

// RemoteCallback 发送远程存储策略上传回调请求
func RemoteCallback(url string, body serializer.UploadCallback) error {
	callbackBody, err := json.Marshal(body)
	if err != nil {
		return serializer.NewError(serializer.CodeCallbackError, "无法编码回调正文", err)
	}

	resp := generalClient.Request(
		"POST",
		url,
		bytes.NewReader(callbackBody),
		WithTimeout(time.Duration(conf.SlaveConfig.CallbackTimeout)*time.Second),
		WithCredential(auth.General, int64(conf.SlaveConfig.SignatureTTL)),
	)

	if resp.Err != nil {
		return serializer.NewError(serializer.CodeCallbackError, "无法发起回调请求", resp.Err)
	}

	// 检查返回HTTP状态码
	if resp.Response.StatusCode != 200 {
		util.Log().Debug("服务端返回非正常状态码：%d", resp.Response.StatusCode)
		return serializer.NewError(serializer.CodeCallbackError, "服务端返回非正常状态码", nil)
	}

	// 检查返回API状态码
	var response serializer.Response
	rawResp, err := ioutil.ReadAll(resp.Response.Body)
	if err != nil {
		return serializer.NewError(serializer.CodeCallbackError, "无法读取响应正文", err)
	}

	// 解析回调服务端响应
	err = json.Unmarshal(rawResp, &response)
	if err != nil {
		util.Log().Debug("无法解析回调服务端响应：%s", string(rawResp))
		return serializer.NewError(serializer.CodeCallbackError, "无法解析服务端返回的响应", err)
	}

	if response.Code != 0 {
		return serializer.NewError(response.Code, response.Msg, errors.New(response.Error))
	}

	return nil
}
