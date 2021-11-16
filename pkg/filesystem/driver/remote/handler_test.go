package remote

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
)

func TestHandler_Token(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{
		Policy: &model.Policy{
			MaxSize:      10,
			AutoRename:   true,
			DirNameRule:  "dir",
			FileNameRule: "file",
			OptionsSerialized: model.PolicyOption{
				FileType: []string{"txt"},
			},
			Server: "http://test.com",
		},
		AuthInstance: auth.HMACAuth{},
	}
	ctx := context.WithValue(context.Background(), fsctx.DisableOverwrite, true)
	auth.General = auth.HMACAuth{SecretKey: []byte("test")}

	// 成功
	{
		cache.Set("setting_siteURL", "http://test.cloudreve.org", 0)
		credential, err := handler.Token(ctx, 10, "123")
		asserts.NoError(err)
		policy, err := serializer.DecodeUploadPolicy(credential.Policy)
		asserts.NoError(err)
		asserts.Equal("http://test.cloudreve.org/api/v3/callback/remote/123", policy.CallbackURL)
		asserts.Equal(uint64(10), policy.MaxSize)
		asserts.Equal(true, policy.AutoRename)
		asserts.Equal("dir", policy.SavePath)
		asserts.Equal("file", policy.FileName)
		asserts.Equal("file", policy.FileName)
		asserts.Equal([]string{"txt"}, policy.AllowedExtension)
	}

}

func TestHandler_Source(t *testing.T) {
	asserts := assert.New(t)
	auth.General = auth.HMACAuth{SecretKey: []byte("test")}

	// 无法获取上下文
	{
		handler := Driver{
			Policy:       &model.Policy{Server: "/"},
			AuthInstance: auth.HMACAuth{},
		}
		ctx := context.Background()
		res, err := handler.Source(ctx, "", url.URL{}, 0, true, 0)
		asserts.NoError(err)
		asserts.NotEmpty(res)
	}

	// 成功
	{
		handler := Driver{
			Policy:       &model.Policy{Server: "/"},
			AuthInstance: auth.HMACAuth{},
		}
		file := model.File{
			SourceName: "1.txt",
		}
		ctx := context.WithValue(context.Background(), fsctx.FileModelCtx, file)
		res, err := handler.Source(ctx, "", url.URL{}, 10, true, 0)
		asserts.NoError(err)
		asserts.Contains(res, "api/v3/slave/download/0")
	}

	// 成功 自定义CDN
	{
		handler := Driver{
			Policy:       &model.Policy{Server: "/", BaseURL: "https://cqu.edu.cn"},
			AuthInstance: auth.HMACAuth{},
		}
		file := model.File{
			SourceName: "1.txt",
		}
		ctx := context.WithValue(context.Background(), fsctx.FileModelCtx, file)
		res, err := handler.Source(ctx, "", url.URL{}, 10, true, 0)
		asserts.NoError(err)
		asserts.Contains(res, "api/v3/slave/download/0")
		asserts.Contains(res, "https://cqu.edu.cn")
	}

	// 解析失败 自定义CDN
	{
		handler := Driver{
			Policy:       &model.Policy{Server: "/", BaseURL: string([]byte{0x7f})},
			AuthInstance: auth.HMACAuth{},
		}
		file := model.File{
			SourceName: "1.txt",
		}
		ctx := context.WithValue(context.Background(), fsctx.FileModelCtx, file)
		res, err := handler.Source(ctx, "", url.URL{}, 10, true, 0)
		asserts.Error(err)
		asserts.Empty(res)
	}

	// 成功 预览
	{
		handler := Driver{
			Policy:       &model.Policy{Server: "/"},
			AuthInstance: auth.HMACAuth{},
		}
		file := model.File{
			SourceName: "1.txt",
		}
		ctx := context.WithValue(context.Background(), fsctx.FileModelCtx, file)
		res, err := handler.Source(ctx, "", url.URL{}, 10, false, 0)
		asserts.NoError(err)
		asserts.Contains(res, "api/v3/slave/source/0")
	}
}

type ClientMock struct {
	testMock.Mock
}

func (m ClientMock) Request(method, target string, body io.Reader, opts ...request.Option) *request.Response {
	args := m.Called(method, target, body, opts)
	return args.Get(0).(*request.Response)
}

func TestHandler_Delete(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{
		Policy: &model.Policy{
			SecretKey: "test",
			Server:    "http://test.com",
		},
		AuthInstance: auth.HMACAuth{},
	}
	ctx := context.Background()
	cache.Set("setting_slave_api_timeout", "60", 0)

	// 成功
	{
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"POST",
			"http://test.com/api/v3/slave/delete",
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`{"code":0}`)),
			},
		})
		handler.Client = clientMock
		failed, err := handler.Delete(ctx, []string{"/test1.txt", "test2.txt"})
		clientMock.AssertExpectations(t)
		asserts.NoError(err)
		asserts.Len(failed, 0)

	}

	// 结果解析失败
	{
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"POST",
			"http://test.com/api/v3/slave/delete",
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`{"code":203}`)),
			},
		})
		handler.Client = clientMock
		failed, err := handler.Delete(ctx, []string{"/test1.txt", "test2.txt"})
		clientMock.AssertExpectations(t)
		asserts.Error(err)
		asserts.Len(failed, 2)
	}

	// 一个失败
	{
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"POST",
			"http://test.com/api/v3/slave/delete",
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`{"code":203,"data":"{\"files\":[\"1\"]}"}`)),
			},
		})
		handler.Client = clientMock
		failed, err := handler.Delete(ctx, []string{"/test1.txt", "test2.txt"})
		clientMock.AssertExpectations(t)
		asserts.Error(err)
		asserts.Len(failed, 1)
	}
}

func TestDriver_List(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{
		Policy: &model.Policy{
			SecretKey: "test",
			Server:    "http://test.com",
		},
		AuthInstance: auth.HMACAuth{},
	}
	ctx := context.Background()
	cache.Set("setting_slave_api_timeout", "60", 0)

	// 成功
	{
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"POST",
			"http://test.com/api/v3/slave/list",
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`{"code":0,"data":"[{}]"}`)),
			},
		})
		handler.Client = clientMock
		res, err := handler.List(ctx, "/", true)
		clientMock.AssertExpectations(t)
		asserts.NoError(err)
		asserts.Len(res, 1)

	}

	// 响应解析失败
	{
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"POST",
			"http://test.com/api/v3/slave/list",
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`{"code":0,"data":"233"}`)),
			},
		})
		handler.Client = clientMock
		res, err := handler.List(ctx, "/", true)
		clientMock.AssertExpectations(t)
		asserts.Error(err)
		asserts.Len(res, 0)
	}

	// 从机返回错误
	{
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"POST",
			"http://test.com/api/v3/slave/list",
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`{"code":203}`)),
			},
		})
		handler.Client = clientMock
		res, err := handler.List(ctx, "/", true)
		clientMock.AssertExpectations(t)
		asserts.Error(err)
		asserts.Len(res, 0)
	}
}

func TestHandler_Get(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{
		Policy: &model.Policy{
			SecretKey: "test",
			Server:    "http://test.com",
		},
		AuthInstance: auth.HMACAuth{},
	}
	ctx := context.Background()

	// 成功
	{
		ctx = context.WithValue(ctx, fsctx.UserCtx, model.User{})
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"GET",
			testMock.Anything,
			nil,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`{"code":0}`)),
			},
		})
		handler.Client = clientMock
		resp, err := handler.Get(ctx, "/test.txt")
		clientMock.AssertExpectations(t)
		asserts.NotNil(resp)
		asserts.NoError(err)
	}

	// 请求失败
	{
		ctx = context.WithValue(ctx, fsctx.UserCtx, model.User{})
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"GET",
			testMock.Anything,
			nil,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 404,
				Body:       ioutil.NopCloser(strings.NewReader(`{"code":0}`)),
			},
		})
		handler.Client = clientMock
		resp, err := handler.Get(ctx, "/test.txt")
		clientMock.AssertExpectations(t)
		asserts.Nil(resp)
		asserts.Error(err)
	}
}

func TestHandler_Put(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{
		Policy: &model.Policy{
			Type:      "remote",
			SecretKey: "test",
			Server:    "http://test.com",
		},
		AuthInstance: auth.HMACAuth{},
	}
	ctx := context.Background()
	asserts.NoError(cache.Set("setting_upload_credential_timeout", "3600", 0))

	// 成功
	{
		ctx = context.WithValue(ctx, fsctx.UserCtx, model.User{})
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"POST",
			"http://test.com/api/v3/slave/upload",
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`{"code":0}`)),
			},
		})
		handler.Client = clientMock
		err := handler.Put(ctx, ioutil.NopCloser(strings.NewReader("test input file")), "/", 15)
		clientMock.AssertExpectations(t)
		asserts.NoError(err)
	}

	// 请求失败
	{
		ctx = context.WithValue(ctx, fsctx.UserCtx, model.User{})
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"POST",
			"http://test.com/api/v3/slave/upload",
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 404,
				Body:       ioutil.NopCloser(strings.NewReader(`{"code":0}`)),
			},
		})
		handler.Client = clientMock
		err := handler.Put(ctx, ioutil.NopCloser(strings.NewReader("test input file")), "/", 15)
		clientMock.AssertExpectations(t)
		asserts.Error(err)
	}

	// 返回错误
	{
		ctx = context.WithValue(ctx, fsctx.UserCtx, model.User{})
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"POST",
			"http://test.com/api/v3/slave/upload",
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`{"code":1}`)),
			},
		})
		handler.Client = clientMock
		err := handler.Put(ctx, ioutil.NopCloser(strings.NewReader("test input file")), "/", 15)
		clientMock.AssertExpectations(t)
		asserts.Error(err)
	}

}

func TestHandler_Thumb(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{
		Policy: &model.Policy{
			Type:      "remote",
			SecretKey: "test",
			Server:    "http://test.com",
		},
		AuthInstance: auth.HMACAuth{},
	}
	ctx := context.Background()
	asserts.NoError(cache.Set("setting_preview_timeout", "60", 0))
	resp, err := handler.Thumb(ctx, "/1.txt")
	asserts.NoError(err)
	asserts.True(resp.Redirect)
}
