package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"sync"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

// GeneralClient 通用 HTTP Client
var GeneralClient Client = NewClient()

// Response 请求的响应或错误信息
type Response struct {
	Err      error
	Response *http.Response
}

// Client 请求客户端
type Client interface {
	Request(method, target string, body io.Reader, opts ...Option) *Response
}

// HTTPClient 实现 Client 接口
type HTTPClient struct {
	mu      sync.Mutex
	options *options
}

func NewClient(opts ...Option) Client {
	client := &HTTPClient{
		options: newDefaultOption(),
	}

	for _, o := range opts {
		o.apply(client.options)
	}

	return client
}

// Request 发送HTTP请求
func (c HTTPClient) Request(method, target string, body io.Reader, opts ...Option) *Response {
	// 应用额外设置
	c.mu.Lock()
	options := *c.options
	c.mu.Unlock()
	for _, o := range opts {
		o.apply(&options)
	}

	// 创建请求客户端
	client := &http.Client{Timeout: options.timeout}

	// size为0时将body设为nil
	if options.contentLength == 0 {
		body = nil
	}

	// 确定请求URL
	if options.endpoint != nil {
		targetURL := *options.endpoint
		targetURL.Path = path.Join(targetURL.Path, target)
		target = targetURL.String()
	}

	// 创建请求
	var (
		req *http.Request
		err error
	)
	if options.ctx != nil {
		req, err = http.NewRequestWithContext(options.ctx, method, target, body)
	} else {
		req, err = http.NewRequest(method, target, body)
	}
	if err != nil {
		return &Response{Err: err}
	}

	// 添加请求相关设置
	if options.header != nil {
		for k, v := range options.header {
			req.Header.Add(k, strings.Join(v, " "))
		}
	}

	if options.masterMeta && conf.SystemConfig.Mode == "master" {
		req.Header.Add("X-Cr-Site-Url", model.GetSiteURL().String())
		req.Header.Add("X-Cr-Site-Id", model.GetSettingByName("siteID"))
		req.Header.Add("X-Cr-Cloudreve-Version", conf.BackendVersion)
	}

	if options.slaveNodeID != "" && conf.SystemConfig.Mode == "slave" {
		req.Header.Add("X-Cr-Node-Id", options.slaveNodeID)
	}

	if options.contentLength != -1 {
		req.ContentLength = options.contentLength
	}

	// 签名请求
	if options.sign != nil {
		switch method {
		case "PUT", "POST", "PATCH":
			auth.SignRequest(options.sign, req, options.signTTL)
		default:
			if resURL, err := auth.SignURI(options.sign, req.URL.String(), options.signTTL); err == nil {
				req.URL = resURL
			}
		}
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return &Response{Err: err}
	}

	return &Response{Err: nil, Response: resp}
}

// GetResponse 检查响应并获取响应正文
func (resp *Response) GetResponse() (string, error) {
	if resp.Err != nil {
		return "", resp.Err
	}
	respBody, err := ioutil.ReadAll(resp.Response.Body)
	_ = resp.Response.Body.Close()

	return string(respBody), err
}

// CheckHTTPResponse 检查请求响应HTTP状态码
func (resp *Response) CheckHTTPResponse(status int) *Response {
	if resp.Err != nil {
		return resp
	}

	// 检查HTTP状态码
	if resp.Response.StatusCode != status {
		resp.Err = fmt.Errorf("服务器返回非正常HTTP状态%d", resp.Response.StatusCode)
	}
	return resp
}

// DecodeResponse 尝试解析为serializer.Response，并对状态码进行检查
func (resp *Response) DecodeResponse() (*serializer.Response, error) {
	if resp.Err != nil {
		return nil, resp.Err
	}

	respString, err := resp.GetResponse()
	if err != nil {
		return nil, err
	}

	var res serializer.Response
	err = json.Unmarshal([]byte(respString), &res)
	if err != nil {
		util.Log().Debug("无法解析回调服务端响应：%s", string(respString))
		return nil, err
	}
	return &res, nil
}

// NopRSCloser 实现不完整seeker
type NopRSCloser struct {
	body   io.ReadCloser
	status *rscStatus
}

type rscStatus struct {
	// http.ServeContent 会读取一小块以决定内容类型，
	// 但是响应body无法实现seek，所以此项为真时第一个read会返回假数据
	IgnoreFirst bool

	Size int64
}

// GetRSCloser 返回带有空seeker的RSCloser，供http.ServeContent使用
func (resp *Response) GetRSCloser() (*NopRSCloser, error) {
	if resp.Err != nil {
		return nil, resp.Err
	}

	return &NopRSCloser{
		body: resp.Response.Body,
		status: &rscStatus{
			Size: resp.Response.ContentLength,
		},
	}, resp.Err
}

// SetFirstFakeChunk 开启第一次read返回空数据
// TODO 测试
func (instance NopRSCloser) SetFirstFakeChunk() {
	instance.status.IgnoreFirst = true
}

// SetContentLength 设置数据流大小
func (instance NopRSCloser) SetContentLength(size int64) {
	instance.status.Size = size
}

// Read 实现 NopRSCloser reader
func (instance NopRSCloser) Read(p []byte) (n int, err error) {
	if instance.status.IgnoreFirst && len(p) == 512 {
		return 0, io.EOF
	}
	return instance.body.Read(p)
}

// Close 实现 NopRSCloser closer
func (instance NopRSCloser) Close() error {
	return instance.body.Close()
}

// Seek 实现 NopRSCloser seeker, 只实现seek开头/结尾以便http.ServeContent用于确定正文大小
func (instance NopRSCloser) Seek(offset int64, whence int) (int64, error) {
	// 进行第一次Seek操作后，取消忽略选项
	if instance.status.IgnoreFirst {
		instance.status.IgnoreFirst = false
	}
	if offset == 0 {
		switch whence {
		case io.SeekStart:
			return 0, nil
		case io.SeekEnd:
			return instance.status.Size, nil
		}
	}
	return 0, errors.New("未实现")

}

// BlackHole 将客户端发来的数据放入黑洞
func BlackHole(r io.Reader) {
	if !model.IsTrueVal(model.GetSettingByName("reset_after_upload_failed")) {
		_, err := io.Copy(ioutil.Discard, r)
		if err != nil {
			util.Log().Debug("黑洞数据出错，%s", err)
		}
	}
}
