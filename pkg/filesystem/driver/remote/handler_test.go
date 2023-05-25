package remote

import (
	"context"
	"errors"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver"
	"github.com/cloudreve/Cloudreve/v3/pkg/mocks/remoteclientmock"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
)

func TestNewDriver(t *testing.T) {
	a := assert.New(t)

	// remoteClient 初始化失败
	{
		d, err := NewDriver(&model.Policy{Server: string([]byte{0x7f})})
		a.Error(err)
		a.Nil(d)
	}

	// 成功
	{
		d, err := NewDriver(&model.Policy{})
		a.NoError(err)
		a.NotNil(d)
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
		res, err := handler.Source(ctx, "", 0, true, 0)
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
		res, err := handler.Source(ctx, "", 10, true, 0)
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
		res, err := handler.Source(ctx, "", 10, true, 0)
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
		res, err := handler.Source(ctx, "", 10, true, 0)
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
		res, err := handler.Source(ctx, "", 10, false, 0)
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
	a := assert.New(t)
	handler, _ := NewDriver(&model.Policy{
		Type:      "remote",
		SecretKey: "test",
		Server:    "http://test.com",
	})
	clientMock := &remoteclientmock.RemoteClientMock{}
	handler.uploadClient = clientMock
	clientMock.On("Upload", testMock.Anything, testMock.Anything).Return(errors.New("error"))
	a.Error(handler.Put(context.Background(), &fsctx.FileStream{}))
	clientMock.AssertExpectations(t)
}

func TestHandler_Thumb(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{
		Policy: &model.Policy{
			Type:      "remote",
			SecretKey: "test",
			Server:    "http://test.com",
			OptionsSerialized: model.PolicyOption{
				ThumbExts: []string{"txt"},
			},
		},
		AuthInstance: auth.HMACAuth{},
	}
	file := &model.File{
		Name:       "1.txt",
		SourceName: "1.txt",
	}
	ctx := context.Background()
	asserts.NoError(cache.Set("setting_preview_timeout", "60", 0))

	// no error
	{
		resp, err := handler.Thumb(ctx, file)
		asserts.NoError(err)
		asserts.True(resp.Redirect)
	}

	// ext not support
	{
		file.Name = "1.jpg"
		resp, err := handler.Thumb(ctx, file)
		asserts.ErrorIs(err, driver.ErrorThumbNotSupported)
		asserts.Nil(resp)
	}
}

func TestHandler_Token(t *testing.T) {
	a := assert.New(t)
	handler, _ := NewDriver(&model.Policy{})

	// 无法创建上传会话
	{
		clientMock := &remoteclientmock.RemoteClientMock{}
		handler.uploadClient = clientMock
		clientMock.On("CreateUploadSession", testMock.Anything, testMock.Anything, int64(10), false).Return(errors.New("error"))
		res, err := handler.Token(context.Background(), 10, &serializer.UploadSession{}, &fsctx.FileStream{})
		a.Error(err)
		a.Contains(err.Error(), "error")
		a.Nil(res)
		clientMock.AssertExpectations(t)
	}

	// 无法创建上传地址
	{
		clientMock := &remoteclientmock.RemoteClientMock{}
		handler.uploadClient = clientMock
		clientMock.On("CreateUploadSession", testMock.Anything, testMock.Anything, int64(10), false).Return(nil)
		clientMock.On("GetUploadURL", int64(10), "").Return("", "", errors.New("error"))
		res, err := handler.Token(context.Background(), 10, &serializer.UploadSession{}, &fsctx.FileStream{})
		a.Error(err)
		a.Contains(err.Error(), "error")
		a.Nil(res)
		clientMock.AssertExpectations(t)
	}

	// 成功
	{
		clientMock := &remoteclientmock.RemoteClientMock{}
		handler.uploadClient = clientMock
		clientMock.On("CreateUploadSession", testMock.Anything, testMock.Anything, int64(10), false).Return(nil)
		clientMock.On("GetUploadURL", int64(10), "").Return("1", "2", nil)
		res, err := handler.Token(context.Background(), 10, &serializer.UploadSession{}, &fsctx.FileStream{})
		a.NoError(err)
		a.NotNil(res)
		a.Equal("1", res.UploadURLs[0])
		a.Equal("2", res.Credential)
		clientMock.AssertExpectations(t)
	}
}

func TestDriver_CancelToken(t *testing.T) {
	a := assert.New(t)
	handler, _ := NewDriver(&model.Policy{})

	clientMock := &remoteclientmock.RemoteClientMock{}
	handler.uploadClient = clientMock
	clientMock.On("DeleteUploadSession", testMock.Anything, "key").Return(errors.New("error"))
	err := handler.CancelToken(context.Background(), &serializer.UploadSession{Key: "key"})
	a.Error(err)
	a.Contains(err.Error(), "error")
	clientMock.AssertExpectations(t)
}
