package request

import (
	"context"
	"errors"
	"fmt"
	"github.com/HFO4/cloudreve/pkg/auth"
	"github.com/HFO4/cloudreve/pkg/filesystem/response"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// GeneralClient 通用 HTTP Client
var GeneralClient Client = HTTPClient{}

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
}

// Option 发送请求的额外设置
type Option interface {
	apply(*options)
}

type options struct {
	timeout time.Duration
	header  http.Header
	sign    auth.Auth
	signTTL int64
	ctx     context.Context
}

type optionFunc func(*options)

func (f optionFunc) apply(o *options) {
	f(o)
}

func newDefaultOption() *options {
	return &options{
		header:  http.Header{},
		timeout: time.Duration(30) * time.Second,
	}
}

// WithTimeout 设置请求超时
func WithTimeout(t time.Duration) Option {
	return optionFunc(func(o *options) {
		o.timeout = t
	})
}

// WithContext 设置请求上下文
func WithContext(c context.Context) Option {
	return optionFunc(func(o *options) {
		o.ctx = c
	})
}

// WithCredential 对请求进行签名
func WithCredential(instance auth.Auth, ttl int64) Option {
	return optionFunc(func(o *options) {
		o.sign = instance
		o.signTTL = ttl
	})
}

// WithHeader 设置请求Header
func WithHeader(header http.Header) Option {
	return optionFunc(func(o *options) {
		o.header = header
	})
}

// Request 发送HTTP请求
func (c HTTPClient) Request(method, target string, body io.Reader, opts ...Option) *Response {
	// 应用额外设置
	options := newDefaultOption()
	for _, o := range opts {
		o.apply(options)
	}

	// 创建请求客户端
	client := &http.Client{Timeout: options.timeout}

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

	// 添加请求header
	req.Header = options.header

	// 签名请求
	if options.sign != nil {
		auth.SignRequest(options.sign, req, options.signTTL)
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

type nopRSCloser struct {
	body io.ReadCloser
	size int64
}

// GetRSCloser 返回带有空seeker的body reader
func (resp *Response) GetRSCloser() (response.RSCloser, error) {
	if resp.Err != nil {
		return nil, resp.Err
	}

	return nopRSCloser{
		body: resp.Response.Body,
		size: resp.Response.ContentLength,
	}, resp.Err
}

// Read 实现 nopRSCloser reader
func (instance nopRSCloser) Read(p []byte) (n int, err error) {
	return instance.body.Read(p)
}

// 实现 nopRSCloser closer
func (instance nopRSCloser) Close() error {
	return instance.body.Close()
}

// 实现 nopRSCloser seeker, 只实现seek开头/结尾以便http.ServeContent用于确定正文大小
func (instance nopRSCloser) Seek(offset int64, whence int) (int64, error) {
	if offset == 0 {
		switch whence {
		case io.SeekStart:
			return 0, nil
		case io.SeekEnd:
			return instance.size, nil
		}
	}
	return 0, errors.New("未实现")

}
