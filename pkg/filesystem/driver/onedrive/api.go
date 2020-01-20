package onedrive

import (
	"context"
	"encoding/json"
	"github.com/HFO4/cloudreve/pkg/request"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// RespError 接口返回错误
type RespError struct {
	APIError APIError `json:"error"`
}

// APIError 接口返回的错误内容
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// UploadSessionResponse 分片上传会话
type UploadSessionResponse struct {
	DataContext        string   `json:"@odata.context"`
	ExpirationDateTime string   `json:"expirationDateTime"`
	NextExpectedRanges []string `json:"nextExpectedRanges"`
	UploadURL          string   `json:"uploadUrl"`
}

// FileInfo 文件元信息
type FileInfo struct {
	Name            string          `json:"name"`
	Size            uint64          `json:"size"`
	Image           imageInfo       `json:"image"`
	ParentReference parentReference `json:"parentReference"`
}

type imageInfo struct {
	Height int `json:"height"`
	Width  int `json:"width"`
}

type parentReference struct {
	Path string `json:"path"`
	Name string `json:"name"`
	ID   string `json:"id"`
}

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
	res, err := client.request(ctx, "GET", requestURL+"?expand=thumbnails", "")
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

	res, err := client.request(ctx, "POST", requestURL, string(bodyBytes))
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

func sysError(err error) *RespError {
	return &RespError{APIError: APIError{
		Code:    "system",
		Message: err.Error(),
	}}
}

func (client *Client) request(ctx context.Context, method string, url string, body string) (string, *RespError) {

	// 获取凭证
	err := client.UpdateCredential(ctx)
	if err != nil {
		return "", sysError(err)
	}

	// 发送请求
	bodyReader := ioutil.NopCloser(strings.NewReader(body))
	res := client.Request.Request(
		method,
		url,
		bodyReader,
		request.WithContentLength(int64(len(body))),
		request.WithHeader(http.Header{
			"Authorization": {"Bearer " + client.Credential.AccessToken},
			"Content-Type":  {"application/json"},
		}),
		request.WithContext(ctx),
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
	if res.Response.StatusCode != 200 {
		decodeErr = json.Unmarshal([]byte(respBody), &errResp)
		if decodeErr != nil {
			return "", sysError(err)
		}
		return "", &errResp
	}

	return respBody, nil
}
