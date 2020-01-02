package request

import (
	"fmt"
	"github.com/HFO4/cloudreve/pkg/auth"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

var GeneralClient Client = HTTPClient{}

// Response 请求的响应或错误信息
type Response struct {
	Err      error
	Response *http.Response
}

// Client 请求客户端
type Client interface {
	Request(method, target string, body io.Reader, opts ...Option) Response
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
func (c HTTPClient) Request(method, target string, body io.Reader, opts ...Option) Response {
	// 应用额外设置
	options := newDefaultOption()
	for _, o := range opts {
		o.apply(options)
	}

	// 创建请求客户端
	client := &http.Client{Timeout: options.timeout}

	// 创建请求
	req, err := http.NewRequest(method, target, body)
	if err != nil {
		return Response{Err: err}
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
		return Response{Err: err}
	}

	return Response{Err: nil, Response: resp}
}

// GetResponse 检查响应并获取响应正文
// todo 测试
func (resp Response) GetResponse(expectStatus int) (string, error) {
	if resp.Err != nil {
		return "", resp.Err
	}
	respBody, err := ioutil.ReadAll(resp.Response.Body)
	if resp.Response.StatusCode != expectStatus {
		return string(respBody),
			fmt.Errorf("服务器返回非正常HTTP状态%d", resp.Response.StatusCode)
	}
	return string(respBody), err
}
