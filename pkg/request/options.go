package request

import (
	"context"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"net/http"
	"net/url"
	"time"
)

// Option 发送请求的额外设置
type Option interface {
	apply(*options)
}

type options struct {
	timeout       time.Duration
	header        http.Header
	sign          auth.Auth
	signTTL       int64
	ctx           context.Context
	contentLength int64
	masterMeta    bool
	endpoint      *url.URL
	slaveNodeID   string
}

type optionFunc func(*options)

func (f optionFunc) apply(o *options) {
	f(o)
}

func newDefaultOption() *options {
	return &options{
		header:        http.Header{},
		timeout:       time.Duration(30) * time.Second,
		contentLength: -1,
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
		for k, v := range header {
			o.header[k] = v
		}
	})
}

// WithoutHeader 设置清除请求Header
func WithoutHeader(header []string) Option {
	return optionFunc(func(o *options) {
		for _, v := range header {
			delete(o.header, v)
		}

	})
}

// WithContentLength 设置请求大小
func WithContentLength(s int64) Option {
	return optionFunc(func(o *options) {
		o.contentLength = s
	})
}

// WithMasterMeta 请求时携带主机信息
func WithMasterMeta() Option {
	return optionFunc(func(o *options) {
		o.masterMeta = true
	})
}

// WithSlaveMeta 请求时携带从机信息
func WithSlaveMeta(s string) Option {
	return optionFunc(func(o *options) {
		o.slaveNodeID = s
	})
}

// Endpoint 使用同一的请求Endpoint
func WithEndpoint(endpoint string) Option {
	endpointURL, _ := url.Parse(endpoint)
	return optionFunc(func(o *options) {
		o.endpoint = endpointURL
	})
}
