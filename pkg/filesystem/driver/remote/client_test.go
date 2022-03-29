package remote

import (
	"context"
	"errors"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/mocks/requestmock"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	a := assert.New(t)
	policy := &model.Policy{}

	// 无法解析服务端url
	{
		policy.Server = string([]byte{0x7f})
		c, err := NewClient(policy)
		a.Error(err)
		a.Nil(c)
	}

	// 成功
	{
		policy.Server = ""
		c, err := NewClient(policy)
		a.NoError(err)
		a.NotNil(c)
	}
}

func TestRemoteClient_Upload(t *testing.T) {
	a := assert.New(t)
	c, _ := NewClient(&model.Policy{})

	// 无法创建上传会话
	{
		clientMock := requestmock.RequestMock{}
		c.(*remoteClient).httpClient = &clientMock
		clientMock.On(
			"Request",
			"PUT",
			"upload",
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: errors.New("error"),
		})
		err := c.Upload(context.Background(), &fsctx.FileStream{})
		a.Error(err)
		a.Contains(err.Error(), "error")
		clientMock.AssertExpectations(t)
	}

	// 分片上传失败，成功删除上传会话
	{
		cache.Set("setting_chunk_retries", "1", 0)
		clientMock := requestmock.RequestMock{}
		c.(*remoteClient).httpClient = &clientMock
		clientMock.On(
			"Request",
			"PUT",
			"upload",
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`{"code":0}`)),
			},
		})
		clientMock.On(
			"Request",
			"POST",
			testMock.Anything,
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: errors.New("error"),
		})
		clientMock.On(
			"Request",
			"DELETE",
			testMock.Anything,
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`{"code":0}`)),
			},
		})
		err := c.Upload(context.Background(), &fsctx.FileStream{})
		a.Error(err)
		a.Contains(err.Error(), "error")
		clientMock.AssertExpectations(t)
	}

	// 分片上传失败，无法删除上传会话
	{
		cache.Set("setting_chunk_retries", "1", 0)
		clientMock := requestmock.RequestMock{}
		c.(*remoteClient).httpClient = &clientMock
		clientMock.On(
			"Request",
			"PUT",
			"upload",
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`{"code":0}`)),
			},
		})
		clientMock.On(
			"Request",
			"POST",
			testMock.Anything,
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: errors.New("error"),
		})
		clientMock.On(
			"Request",
			"DELETE",
			testMock.Anything,
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: errors.New("error2"),
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`{"code":0}`)),
			},
		})
		err := c.Upload(context.Background(), &fsctx.FileStream{})
		a.Error(err)
		a.Contains(err.Error(), "error")
		clientMock.AssertExpectations(t)
	}

	// 成功
	{
		cache.Set("setting_chunk_retries", "1", 0)
		clientMock := requestmock.RequestMock{}
		c.(*remoteClient).httpClient = &clientMock
		clientMock.On(
			"Request",
			"PUT",
			"upload",
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`{"code":0}`)),
			},
		})
		clientMock.On(
			"Request",
			"POST",
			testMock.Anything,
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`{"code":0}`)),
			},
		})
		err := c.Upload(context.Background(), &fsctx.FileStream{})
		a.NoError(err)
		clientMock.AssertExpectations(t)
	}
}

func TestRemoteClient_CreateUploadSessionFailed(t *testing.T) {
	a := assert.New(t)
	c, _ := NewClient(&model.Policy{})

	clientMock := requestmock.RequestMock{}
	c.(*remoteClient).httpClient = &clientMock
	clientMock.On(
		"Request",
		"PUT",
		"upload",
		testMock.Anything,
		testMock.Anything,
	).Return(&request.Response{
		Err: nil,
		Response: &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(strings.NewReader(`{"code":500,"msg":"error"}`)),
		},
	})
	err := c.Upload(context.Background(), &fsctx.FileStream{})
	a.Error(err)
	a.Contains(err.Error(), "error")
	clientMock.AssertExpectations(t)
}

func TestRemoteClient_UploadChunkFailed(t *testing.T) {
	a := assert.New(t)
	c, _ := NewClient(&model.Policy{})

	clientMock := requestmock.RequestMock{}
	c.(*remoteClient).httpClient = &clientMock
	clientMock.On(
		"Request",
		"POST",
		testMock.Anything,
		testMock.Anything,
		testMock.Anything,
	).Return(&request.Response{
		Err: nil,
		Response: &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(strings.NewReader(`{"code":500,"msg":"error"}`)),
		},
	})
	err := c.(*remoteClient).uploadChunk(context.Background(), "", 0, strings.NewReader(""), false, 0)
	a.Error(err)
	a.Contains(err.Error(), "error")
	clientMock.AssertExpectations(t)
}

func TestRemoteClient_GetUploadURL(t *testing.T) {
	a := assert.New(t)
	c, _ := NewClient(&model.Policy{})

	// url 解析失败
	{
		c.(*remoteClient).policy.Server = string([]byte{0x7f})
		res, sign, err := c.GetUploadURL(0, "")
		a.Error(err)
		a.Empty(res)
		a.Empty(sign)
	}

	// 成功
	{
		c.(*remoteClient).policy.Server = ""
		res, sign, err := c.GetUploadURL(0, "")
		a.NoError(err)
		a.NotEmpty(res)
		a.NotEmpty(sign)
	}
}
