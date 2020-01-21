package onedrive

import (
	"context"
	"encoding/json"
	"errors"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/cache"
	"github.com/HFO4/cloudreve/pkg/request"
	"github.com/HFO4/cloudreve/pkg/util"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

// GetSourcePath 获取文件的绝对路径
func (info *FileInfo) GetSourcePath() string {
	res, err := url.PathUnescape(
		strings.TrimPrefix(
			path.Join(
				strings.TrimPrefix(info.ParentReference.Path, "/drive/root:"),
				info.Name,
			),
			"/",
		),
	)
	if err != nil {
		return ""
	}
	return res
}

// Error 实现error接口
func (err RespError) Error() string {
	return err.APIError.Message
}

func (client *Client) getRequestURL(api string) string {
	base, _ := url.Parse(client.Endpoints.EndpointURL)
	if base == nil {
		return ""
	}
	base.Path = path.Join(base.Path, api)
	return base.String()
}

// Meta 根据资源ID获取文件元信息
func (client *Client) Meta(ctx context.Context, id string) (*FileInfo, error) {

	requestURL := client.getRequestURL("/me/drive/items/" + id)
	res, err := client.requestWithStr(ctx, "GET", requestURL+"?expand=thumbnails", "", 200)
	if err != nil {
		return nil, err
	}

	var (
		decodeErr error
		fileInfo  FileInfo
	)
	decodeErr = json.Unmarshal([]byte(res), &fileInfo)
	if decodeErr != nil {
		return nil, decodeErr
	}

	return &fileInfo, nil

}

// CreateUploadSession 创建分片上传会话
func (client *Client) CreateUploadSession(ctx context.Context, dst string, opts ...Option) (string, error) {

	options := newDefaultOption()
	for _, o := range opts {
		o.apply(options)
	}

	dst = strings.TrimPrefix(dst, "/")
	requestURL := client.getRequestURL("me/drive/root:/" + dst + ":/createUploadSession")
	body := map[string]map[string]interface{}{
		"item": {
			"@microsoft.graph.conflictBehavior": options.conflictBehavior,
		},
	}
	bodyBytes, _ := json.Marshal(body)

	res, err := client.requestWithStr(ctx, "POST", requestURL, string(bodyBytes), 200)
	if err != nil {
		return "", err
	}

	var (
		decodeErr     error
		uploadSession UploadSessionResponse
	)
	decodeErr = json.Unmarshal([]byte(res), &uploadSession)
	if decodeErr != nil {
		return "", decodeErr
	}

	return uploadSession.UploadURL, nil
}

// GetUploadSessionStatus 查询上传会话状态
func (client *Client) GetUploadSessionStatus(ctx context.Context, uploadURL string) (*UploadSessionResponse, error) {
	res, err := client.requestWithStr(ctx, "GET", uploadURL, "", 200)
	if err != nil {
		return nil, err
	}

	var (
		decodeErr     error
		uploadSession UploadSessionResponse
	)
	decodeErr = json.Unmarshal([]byte(res), &uploadSession)
	if decodeErr != nil {
		return nil, decodeErr
	}

	return &uploadSession, nil
}

// DeleteUploadSession 删除上传会话
func (client *Client) DeleteUploadSession(ctx context.Context, uploadURL string) error {
	_, err := client.requestWithStr(ctx, "DELETE", uploadURL, "", 204)
	if err != nil {
		return err
	}

	return nil
}

// PutFile 上传小文件到dst
func (client *Client) PutFile(ctx context.Context, dst string, body io.Reader) (*UploadResult, error) {
	dst = strings.TrimPrefix(dst, "/")
	requestURL := client.getRequestURL("me/drive/root:/" + dst + ":/content")

	res, err := client.request(ctx, "PUT", requestURL, body, 201)
	if err != nil {
		return nil, err
	}

	var (
		decodeErr error
		uploadRes UploadResult
	)
	decodeErr = json.Unmarshal([]byte(res), &uploadRes)
	if decodeErr != nil {
		return nil, decodeErr
	}

	return &uploadRes, nil
}

// Delete 并行删除文件，返回删除失败的文件，及第一个遇到的错误
func (client *Client) Delete(ctx context.Context, dst []string) ([]string, error) {
	body := client.makeBatchDeleteRequestsBody(dst)
	res, err := client.requestWithStr(ctx, "POST", client.getRequestURL("$batch"), body, 200)
	if err != nil {
		return dst, err
	}

	var (
		decodeErr error
		deleteRes BatchResponses
	)
	decodeErr = json.Unmarshal([]byte(res), &deleteRes)
	if decodeErr != nil {
		return dst, decodeErr
	}

	// 取得删除失败的文件
	failed := getDeleteFailed(&deleteRes)
	if len(failed) != 0 {
		return failed, errors.New("无法删除文件")
	}
	return failed, nil
}

func getDeleteFailed(res *BatchResponses) []string {
	var failed = make([]string, 0, len(res.Responses))
	for _, v := range res.Responses {
		if v.Status != 204 {
			failed = append(failed, v.ID)
		}
	}
	return failed
}

// makeBatchDeleteRequestsBody 生成批量删除请求正文
func (client *Client) makeBatchDeleteRequestsBody(files []string) string {
	req := BatchRequests{
		Requests: make([]BatchRequest, len(files)),
	}
	for i, v := range files {
		v = strings.TrimPrefix(v, "/")
		req.Requests[i] = BatchRequest{
			ID:     v,
			Method: "DELETE",
			URL:    "me/drive/root:/" + v,
		}
	}

	res, _ := json.Marshal(req)
	return string(res)
}

// MonitorUpload 监控客户端分片上传进度
func (client *Client) MonitorUpload(uploadURL, callbackKey, path string, size uint64, ttl int64) {
	// 回调完成通知chan
	callbackChan := make(chan bool)
	callbackSignal.Store(callbackKey, callbackChan)
	timeout := model.GetIntSetting("onedrive_monitor_timeout", 600)
	interval := model.GetIntSetting("onedrive_callback_check", 20)

	for {
		select {
		case <-callbackChan:
			util.Log().Debug("客户端完成回调")
			return
		case <-time.After(time.Duration(ttl) * time.Second):
			// 上传会话到期，仍未完成上传，创建占位符
			client.DeleteUploadSession(context.Background(), uploadURL)
			_, err := client.PutFile(context.Background(), path, strings.NewReader(""))
			if err != nil {
				util.Log().Debug("无法创建占位文件，%s", err)
			}
			return
		case <-time.After(time.Duration(timeout) * time.Second):
			util.Log().Debug("检查上传情况")
			status, err := client.GetUploadSessionStatus(context.Background(), uploadURL)

			if err != nil {
				if resErr, ok := err.(*RespError); ok {
					if resErr.APIError.Code == "itemNotFound" {
						util.Log().Debug("上传会话已完成，稍后检查回调")
						time.Sleep(time.Duration(interval) * time.Second)
						util.Log().Debug("开始检查回调")
						_, ok := cache.Get("callback_" + callbackKey)
						if ok {
							util.Log().Warning("未发送回调，删除文件")
							cache.Deletes([]string{callbackKey}, "callback_")
							_, err = client.Delete(context.Background(), []string{path})
							if err != nil {
								util.Log().Warning("无法删除未回掉的文件，%s", err)
							}
						}
						return
					}
					util.Log().Debug("无法获取上传会话状态，%s", err.Error())
					return
				}
				util.Log().Debug("无法获取上传会话状态，继续下一轮，%s", err.Error())
			}

			// 成功获取分片上传状态，检查文件大小
			sizeRange := strings.Split(
				status.NextExpectedRanges[len(status.NextExpectedRanges)-1],
				"-",
			)
			if len(sizeRange) != 2 {
				continue
			}
			uploadFullSize, _ := strconv.ParseUint(sizeRange[1], 10, 64)
			if (sizeRange[0] == "0" && sizeRange[1] == "") || uploadFullSize+1 != size {
				util.Log().Debug("未开始上传或文件大小不一致，取消上传会话")
				// 取消上传会话，实测OneDrive取消上传会话后，客户端还是可以上传，
				// 所以上传一个空文件占位，阻止客户端上传
				client.DeleteUploadSession(context.Background(), uploadURL)
				_, err := client.PutFile(context.Background(), path, strings.NewReader(""))
				if err != nil {
					util.Log().Debug("无法创建占位文件，%s", err)
				}
				return
			}

		}
	}
}

// FinishCallback 向Monitor发送回调结束信号
func FinishCallback(key string) {
	if signal, ok := callbackSignal.Load(key); ok {
		if signalChan, ok := signal.(chan bool); ok {
			close(signalChan)
		}
	}
}

func sysError(err error) *RespError {
	return &RespError{APIError: APIError{
		Code:    "system",
		Message: err.Error(),
	}}
}

func (client *Client) request(ctx context.Context, method string, url string, body io.Reader, expectedCode int, option ...request.Option) (string, *RespError) {
	// 获取凭证
	err := client.UpdateCredential(ctx)
	if err != nil {
		return "", sysError(err)
	}

	option = append(option,
		request.WithHeader(http.Header{
			"Authorization": {"Bearer " + client.Credential.AccessToken},
			"Content-Type":  {"application/json"},
		}),
		request.WithContext(ctx),
	)

	// 发送请求
	res := client.Request.Request(
		method,
		url,
		body,
		option...,
	)

	if res.Err != nil {
		return "", sysError(res.Err)
	}

	respBody, err := res.GetResponse()
	if err != nil {
		return "", sysError(err)
	}

	// 解析请求响应
	var (
		errResp   RespError
		decodeErr error
	)
	// 如果有错误
	if res.Response.StatusCode != expectedCode {
		decodeErr = json.Unmarshal([]byte(respBody), &errResp)
		if decodeErr != nil {
			return "", sysError(err)
		}
		return "", &errResp
	}

	return respBody, nil
}

func (client *Client) requestWithStr(ctx context.Context, method string, url string, body string, expectedCode int) (string, *RespError) {
	// 发送请求
	bodyReader := ioutil.NopCloser(strings.NewReader(body))
	return client.request(ctx, method, url, bodyReader, expectedCode,
		request.WithContentLength(int64(len(body))),
	)
}
