package oss

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
)

func TestDriver_InitOSSClient(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{
		Policy: &model.Policy{
			AccessKey:  "ak",
			SecretKey:  "sk",
			BucketName: "test",
			Server:     "test.com",
		},
	}

	// 成功
	{
		asserts.NoError(handler.InitOSSClient(false))
	}

	// 使用内网Endpoint
	{
		handler.Policy.OptionsSerialized.ServerSideEndpoint = "endpoint2"
		asserts.NoError(handler.InitOSSClient(false))
	}

	// 未指定存储策略
	{
		handler := Driver{}
		asserts.Error(handler.InitOSSClient(false))
	}
}

func TestDriver_CORS(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{
		Policy: &model.Policy{
			AccessKey:  "ak",
			SecretKey:  "sk",
			BucketName: "test",
			Server:     "test.com",
		},
	}

	// 失败
	{
		asserts.NotPanics(func() {
			handler.CORS()
		})
	}
}

func TestDriver_Token(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{
		Policy: &model.Policy{
			AccessKey:  "ak",
			SecretKey:  "sk",
			BucketName: "test",
			Server:     "test.com",
		},
	}

	// 成功
	{
		ctx := context.WithValue(context.Background(), fsctx.SavePathCtx, "/123")
		cache.Set("setting_siteURL", "http://test.cloudreve.org", 0)
		res, err := handler.Token(ctx, 10, "key")
		asserts.NoError(err)
		asserts.NotEmpty(res.Policy)
		asserts.NotEmpty(res.Token)
		asserts.Equal(handler.Policy.AccessKey, res.AccessKey)
		asserts.Equal("/123", res.Path)
	}

	// 上下文错误
	{
		ctx := context.Background()
		_, err := handler.Token(ctx, 10, "key")
		asserts.Error(err)
	}

}

func TestDriver_Source(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{
		Policy: &model.Policy{
			AccessKey:  "ak",
			SecretKey:  "sk",
			BucketName: "test",
			Server:     "test.com",
			IsPrivate:  true,
		},
	}

	// 正常 非下载 无限速
	{
		res, err := handler.Source(context.Background(), "/123", url.URL{}, 10, false, 0)
		asserts.NoError(err)
		resURL, err := url.Parse(res)
		asserts.NoError(err)
		query := resURL.Query()
		asserts.NotEmpty(query.Get("Signature"))
		asserts.NotEmpty(query.Get("Expires"))
		asserts.Equal("ak", query.Get("OSSAccessKeyId"))
	}

	// 限速 + 下载
	{
		ctx := context.WithValue(context.Background(), fsctx.FileModelCtx, model.File{Name: "123.txt"})
		res, err := handler.Source(ctx, "/123", url.URL{}, 10, true, 102401)
		asserts.NoError(err)
		resURL, err := url.Parse(res)
		asserts.NoError(err)
		query := resURL.Query()
		asserts.NotEmpty(query.Get("Signature"))
		asserts.NotEmpty(query.Get("Expires"))
		asserts.Equal("ak", query.Get("OSSAccessKeyId"))
		asserts.EqualValues("819208", query.Get("x-oss-traffic-limit"))
		asserts.NotEmpty(query.Get("response-content-disposition"))
	}

	// 限速超出范围 + 下载
	{
		ctx := context.WithValue(context.Background(), fsctx.FileModelCtx, model.File{Name: "123.txt"})
		res, err := handler.Source(ctx, "/123", url.URL{}, 10, true, 10)
		asserts.NoError(err)
		resURL, err := url.Parse(res)
		asserts.NoError(err)
		query := resURL.Query()
		asserts.NotEmpty(query.Get("Signature"))
		asserts.NotEmpty(query.Get("Expires"))
		asserts.Equal("ak", query.Get("OSSAccessKeyId"))
		asserts.EqualValues("819200", query.Get("x-oss-traffic-limit"))
		asserts.NotEmpty(query.Get("response-content-disposition"))
	}

	// 限速超出范围 + 下载
	{
		ctx := context.WithValue(context.Background(), fsctx.FileModelCtx, model.File{Name: "123.txt"})
		res, err := handler.Source(ctx, "/123", url.URL{}, 10, true, 838860801)
		asserts.NoError(err)
		resURL, err := url.Parse(res)
		asserts.NoError(err)
		query := resURL.Query()
		asserts.NotEmpty(query.Get("Signature"))
		asserts.NotEmpty(query.Get("Expires"))
		asserts.Equal("ak", query.Get("OSSAccessKeyId"))
		asserts.EqualValues("838860800", query.Get("x-oss-traffic-limit"))
		asserts.NotEmpty(query.Get("response-content-disposition"))
	}

	// 公共空间
	{
		handler.Policy.IsPrivate = false
		res, err := handler.Source(context.Background(), "/123", url.URL{}, 10, false, 0)
		asserts.NoError(err)
		resURL, err := url.Parse(res)
		asserts.NoError(err)
		query := resURL.Query()
		asserts.Empty(query.Get("Signature"))
	}

	// 正常 指定了CDN域名
	{
		handler.Policy.BaseURL = "https://cqu.edu.cn"
		res, err := handler.Source(context.Background(), "/123", url.URL{}, 10, false, 0)
		asserts.NoError(err)
		resURL, err := url.Parse(res)
		asserts.NoError(err)
		query := resURL.Query()
		asserts.Empty(query.Get("Signature"))
		asserts.Contains(resURL.String(), handler.Policy.BaseURL)
	}

	// 强制使用公网 Endpoint
	{
		handler.Policy.BaseURL = ""
		handler.Policy.OptionsSerialized.ServerSideEndpoint = "endpoint.com"
		res, err := handler.Source(context.WithValue(context.Background(), fsctx.ForceUsePublicEndpointCtx, false), "/123", url.URL{}, 10, false, 0)
		asserts.NoError(err)
		resURL, err := url.Parse(res)
		asserts.NoError(err)
		query := resURL.Query()
		asserts.Empty(query.Get("Signature"))
		asserts.Contains(resURL.String(), "endpoint.com")
	}
}

func TestDriver_Thumb(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{
		Policy: &model.Policy{
			AccessKey:  "ak",
			SecretKey:  "sk",
			BucketName: "test",
			Server:     "test.com",
		},
	}

	// 上下文不存在
	{
		ctx := context.Background()
		res, err := handler.Thumb(ctx, "/123.txt")
		asserts.Error(err)
		asserts.Nil(res)
	}

	// 成功
	{
		cache.Set("setting_preview_timeout", "60", 0)
		ctx := context.WithValue(context.Background(), fsctx.ThumbSizeCtx, [2]uint{10, 20})
		res, err := handler.Thumb(ctx, "/123.jpg")
		asserts.NoError(err)
		resURL, err := url.Parse(res.URL)
		asserts.NoError(err)
		urlQuery := resURL.Query()
		asserts.Equal("image/resize,m_lfit,h_20,w_10", urlQuery.Get("x-oss-process"))
	}
}

func TestDriver_Delete(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{
		Policy: &model.Policy{
			AccessKey:  "ak",
			SecretKey:  "sk",
			BucketName: "test",
			Server:     "oss-cn-shanghai.aliyuncs.com",
		},
	}

	// 失败
	{
		res, err := handler.Delete(context.Background(), []string{"1", "2", "3"})
		asserts.Error(err)
		asserts.Equal([]string{"1", "2", "3"}, res)
	}
}

func TestDriver_Put(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{
		Policy: &model.Policy{
			AccessKey:  "ak",
			SecretKey:  "sk",
			BucketName: "test",
			Server:     "oss-cn-shanghai.aliyuncs.com",
		},
	}
	cache.Set("setting_upload_credential_timeout", "3600", 0)
	ctx := context.WithValue(context.Background(), fsctx.DisableOverwrite, true)

	// 失败
	{
		err := handler.Put(ctx, ioutil.NopCloser(strings.NewReader("123")), "/123.txt", 3)
		asserts.Error(err)
	}
}

type ClientMock struct {
	testMock.Mock
}

func (m ClientMock) Request(method, target string, body io.Reader, opts ...request.Option) *request.Response {
	args := m.Called(method, target, body, opts)
	return args.Get(0).(*request.Response)
}

func TestDriver_Get(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{
		Policy: &model.Policy{
			AccessKey:  "ak",
			SecretKey:  "sk",
			BucketName: "test",
			Server:     "oss-cn-shanghai.aliyuncs.com",
		},
		HTTPClient: request.NewClient(),
	}
	cache.Set("setting_preview_timeout", "3600", 0)

	// 响应失败
	{
		res, err := handler.Get(context.Background(), "123.txt")
		asserts.Error(err)
		asserts.Nil(res)
	}

	// 响应成功
	{
		ctx := context.WithValue(context.Background(), fsctx.FileModelCtx, model.File{Size: 3})
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"GET",
			testMock.Anything,
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`123`)),
			},
		})
		handler.HTTPClient = clientMock
		res, err := handler.Get(ctx, "123.txt")
		clientMock.AssertExpectations(t)
		asserts.NoError(err)
		n, err := res.Seek(0, io.SeekEnd)
		asserts.NoError(err)
		asserts.EqualValues(3, n)
		content, err := ioutil.ReadAll(res)
		asserts.NoError(err)
		asserts.Equal("123", string(content))
	}
}

func TestDriver_List(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{
		Policy: &model.Policy{
			AccessKey:  "ak",
			SecretKey:  "sk",
			BucketName: "test",
			Server:     "test.com",
			IsPrivate:  true,
		},
	}

	// 连接失败
	{
		res, err := handler.List(context.Background(), "/", true)
		asserts.Error(err)
		asserts.Empty(res)
	}
}
