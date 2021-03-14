package onedrive

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
)

func TestRequest(t *testing.T) {
	asserts := assert.New(t)
	client := Client{
		Policy:   &model.Policy{},
		ClientID: "TestRequest",
		Credential: &Credential{
			ExpiresIn:    time.Now().Add(time.Duration(100) * time.Hour).Unix(),
			AccessToken:  "AccessToken",
			RefreshToken: "RefreshToken",
		},
	}

	// 请求发送失败
	{
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"POST",
			"http://dev.com",
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: errors.New("error"),
		})
		client.Request = clientMock
		res, err := client.request(context.Background(), "POST", "http://dev.com", strings.NewReader(""))
		clientMock.AssertExpectations(t)
		asserts.Error(err)
		asserts.Empty(res)
		asserts.Equal("error", err.Error())
	}

	// 无法更新凭证
	{
		client.Credential.RefreshToken = ""
		client.Credential.AccessToken = ""
		res, err := client.request(context.Background(), "POST", "http://dev.com", strings.NewReader(""))
		asserts.Error(err)
		asserts.Empty(res)
		client.Credential.RefreshToken = "RefreshToken"
		client.Credential.AccessToken = "AccessToken"
	}

	// 无法获取响应正文
	{
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"POST",
			"http://dev.com",
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(mockReader("")),
			},
		})
		client.Request = clientMock
		res, err := client.request(context.Background(), "POST", "http://dev.com", strings.NewReader(""))
		clientMock.AssertExpectations(t)
		asserts.Error(err)
		asserts.Empty(res)
	}

	// OneDrive返回错误
	{
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"POST",
			"http://dev.com",
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 400,
				Body:       ioutil.NopCloser(strings.NewReader(`{"error":{"message":"error msg"}}`)),
			},
		})
		client.Request = clientMock
		res, err := client.request(context.Background(), "POST", "http://dev.com", strings.NewReader(""))
		clientMock.AssertExpectations(t)
		asserts.Error(err)
		asserts.Empty(res)
		asserts.Equal("error msg", err.Error())
	}

	// OneDrive返回未知响应
	{
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"POST",
			"http://dev.com",
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 400,
				Body:       ioutil.NopCloser(strings.NewReader(`???`)),
			},
		})
		client.Request = clientMock
		res, err := client.request(context.Background(), "POST", "http://dev.com", strings.NewReader(""))
		clientMock.AssertExpectations(t)
		asserts.Error(err)
		asserts.Empty(res)
	}
}

func TestFileInfo_GetSourcePath(t *testing.T) {
	asserts := assert.New(t)

	// 成功
	{
		fileInfo := FileInfo{
			Name: "%e6%96%87%e4%bb%b6%e5%90%8d.jpg",
			ParentReference: parentReference{
				Path: "/drive/root:/123/321",
			},
		}
		asserts.Equal("123/321/文件名.jpg", fileInfo.GetSourcePath())
	}

	// 失败
	{
		fileInfo := FileInfo{
			Name: "%e6%96%87%e4%bb%b6%e5%90%8g.jpg",
			ParentReference: parentReference{
				Path: "/drive/root:/123/321",
			},
		}
		asserts.Equal("", fileInfo.GetSourcePath())
	}
}

func TestClient_GetRequestURL(t *testing.T) {
	asserts := assert.New(t)
	client, _ := NewClient(&model.Policy{})

	// 出错
	{
		client.Endpoints.EndpointURL = string([]byte{0x7f})
		asserts.Equal("", client.getRequestURL("123"))
	}

	// 使用DriverResource
	{
		client.Endpoints.EndpointURL = "https://graph.microsoft.com/v1.0"
		asserts.Equal("https://graph.microsoft.com/v1.0/me/drive/123", client.getRequestURL("123"))
	}

	// 不使用DriverResource
	{
		client.Endpoints.EndpointURL = "https://graph.microsoft.com/v1.0"
		asserts.Equal("https://graph.microsoft.com/v1.0/123", client.getRequestURL("123", WithDriverResource(false)))
	}
}

func TestClient_GetSiteIDByURL(t *testing.T) {
	asserts := assert.New(t)
	client, _ := NewClient(&model.Policy{})
	client.Credential.AccessToken = "AccessToken"

	// 请求失败
	{
		client.Credential.ExpiresIn = 0
		res, err := client.GetSiteIDByURL(context.Background(), "https://cquedu.sharepoint.com")
		asserts.Error(err)
		asserts.Empty(res)

	}

	// 返回未知响应
	{
		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
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
				Body:       ioutil.NopCloser(strings.NewReader(`???`)),
			},
		})
		client.Request = clientMock
		res, err := client.GetSiteIDByURL(context.Background(), "https://cquedu.sharepoint.com")
		clientMock.AssertExpectations(t)
		asserts.Error(err)
		asserts.Empty(res)
	}

	// 返回正常
	{
		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
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
				Body:       ioutil.NopCloser(strings.NewReader(`{"id":"123321"}`)),
			},
		})
		client.Request = clientMock
		res, err := client.GetSiteIDByURL(context.Background(), "https://cquedu.sharepoint.com")
		clientMock.AssertExpectations(t)
		asserts.NoError(err)
		asserts.NotEmpty(res)
		asserts.Equal("123321", res)
	}
}

func TestClient_Meta(t *testing.T) {
	asserts := assert.New(t)
	client, _ := NewClient(&model.Policy{})
	client.Credential.AccessToken = "AccessToken"

	// 请求失败
	{
		client.Credential.ExpiresIn = 0
		res, err := client.Meta(context.Background(), "", "123")
		asserts.Error(err)
		asserts.Nil(res)

	}

	// 返回未知响应
	{
		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
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
				Body:       ioutil.NopCloser(strings.NewReader(`???`)),
			},
		})
		client.Request = clientMock
		res, err := client.Meta(context.Background(), "", "123")
		clientMock.AssertExpectations(t)
		asserts.Error(err)
		asserts.Nil(res)
	}

	// 返回正常
	{
		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
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
				Body:       ioutil.NopCloser(strings.NewReader(`{"name":"123321"}`)),
			},
		})
		client.Request = clientMock
		res, err := client.Meta(context.Background(), "", "123")
		clientMock.AssertExpectations(t)
		asserts.NoError(err)
		asserts.NotNil(res)
		asserts.Equal("123321", res.Name)
	}
}

func TestClient_CreateUploadSession(t *testing.T) {
	asserts := assert.New(t)
	client, _ := NewClient(&model.Policy{})
	client.Credential.AccessToken = "AccessToken"

	// 请求失败
	{
		client.Credential.ExpiresIn = 0
		res, err := client.CreateUploadSession(context.Background(), "123.jpg")
		asserts.Error(err)
		asserts.Empty(res)

	}

	// 返回未知响应
	{
		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
		clientMock := ClientMock{}
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
				Body:       ioutil.NopCloser(strings.NewReader(`???`)),
			},
		})
		client.Request = clientMock
		res, err := client.CreateUploadSession(context.Background(), "123.jpg")
		clientMock.AssertExpectations(t)
		asserts.Error(err)
		asserts.Empty(res)
	}

	// 返回正常
	{
		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
		clientMock := ClientMock{}
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
				Body:       ioutil.NopCloser(strings.NewReader(`{"uploadUrl":"123321"}`)),
			},
		})
		client.Request = clientMock
		res, err := client.CreateUploadSession(context.Background(), "123.jpg", WithConflictBehavior("fail"))
		clientMock.AssertExpectations(t)
		asserts.NoError(err)
		asserts.NotNil(res)
		asserts.Equal("123321", res)
	}
}

func TestClient_GetUploadSessionStatus(t *testing.T) {
	asserts := assert.New(t)
	client, _ := NewClient(&model.Policy{})
	client.Credential.AccessToken = "AccessToken"

	// 请求失败
	{
		client.Credential.ExpiresIn = 0
		res, err := client.GetUploadSessionStatus(context.Background(), "http://dev.com")
		asserts.Error(err)
		asserts.Empty(res)

	}

	// 返回未知响应
	{
		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"GET",
			"http://dev.com",
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`???`)),
			},
		})
		client.Request = clientMock
		res, err := client.GetUploadSessionStatus(context.Background(), "http://dev.com")
		clientMock.AssertExpectations(t)
		asserts.Error(err)
		asserts.Nil(res)
	}

	// 返回正常
	{
		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"GET",
			"http://dev.com",
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`{"uploadUrl":"123321"}`)),
			},
		})
		client.Request = clientMock
		res, err := client.GetUploadSessionStatus(context.Background(), "http://dev.com")
		clientMock.AssertExpectations(t)
		asserts.NoError(err)
		asserts.NotNil(res)
		asserts.Equal("123321", res.UploadURL)
	}
}

func TestClient_UploadChunk(t *testing.T) {
	asserts := assert.New(t)
	client, _ := NewClient(&model.Policy{})
	client.Credential.AccessToken = "AccessToken"
	client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()

	// 非最后分片，正常
	{
		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"PUT",
			"http://dev.com",
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`{"uploadUrl":"http://dev.com/2"}`)),
			},
		})
		client.Request = clientMock
		res, err := client.UploadChunk(context.Background(), "http://dev.com", &Chunk{
			Offset:    0,
			ChunkSize: 10,
			Total:     100,
			Retried:   0,
			Data:      []byte("12313121231312"),
		})
		clientMock.AssertExpectations(t)
		asserts.NoError(err)
		asserts.Equal("http://dev.com/2", res.UploadURL)
	}

	// 非最后分片，异常响应
	{
		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"PUT",
			"http://dev.com",
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`???`)),
			},
		})
		client.Request = clientMock
		res, err := client.UploadChunk(context.Background(), "http://dev.com", &Chunk{
			Offset:    0,
			ChunkSize: 10,
			Total:     100,
			Retried:   0,
			Data:      []byte("12313112313122"),
		})
		clientMock.AssertExpectations(t)
		asserts.Error(err)
		asserts.Nil(res)
	}

	// 最后分片，正常
	{
		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"PUT",
			"http://dev.com",
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`???`)),
			},
		})
		client.Request = clientMock
		res, err := client.UploadChunk(context.Background(), "http://dev.com", &Chunk{
			Offset:    95,
			ChunkSize: 5,
			Total:     100,
			Retried:   0,
			Data:      []byte("1231312"),
		})
		clientMock.AssertExpectations(t)
		asserts.NoError(err)
		asserts.Nil(res)
	}

	// 最后分片，第一次失败，重试后成功
	{
		cache.Set("setting_onedrive_chunk_retries", "1", 0)
		client.Credential.ExpiresIn = 0
		go func() {
			time.Sleep(time.Duration(2) * time.Second)
			client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
		}()
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"PUT",
			"http://dev.com",
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`???`)),
			},
		})
		client.Request = clientMock
		chunk := &Chunk{
			Offset:    95,
			ChunkSize: 5,
			Total:     100,
			Retried:   0,
			Data:      []byte("1231312"),
		}
		res, err := client.UploadChunk(context.Background(), "http://dev.com", chunk)
		clientMock.AssertExpectations(t)
		asserts.NoError(err)
		asserts.Nil(res)
		asserts.EqualValues(1, chunk.Retried)
	}
}

func TestClient_Upload(t *testing.T) {
	asserts := assert.New(t)
	client, _ := NewClient(&model.Policy{})
	client.Credential.AccessToken = "AccessToken"
	client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
	ctx := context.WithValue(context.Background(), fsctx.DisableOverwrite, true)

	// 小文件，简单上传，失败
	{
		client.Credential.ExpiresIn = 0
		err := client.Upload(ctx, "123.jpg", 3, strings.NewReader("123"))
		asserts.Error(err)
	}

	// 上下文取消
	{
		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
		clientMock := ClientMock{}
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
				Body:       ioutil.NopCloser(strings.NewReader(`{"uploadUrl":"123321"}`)),
			},
		})
		client.Request = clientMock
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := client.Upload(ctx, "123.jpg", 15*1024*1024, strings.NewReader("123"))
		clientMock.AssertExpectations(t)
		asserts.Error(err)
		asserts.Equal(ErrClientCanceled, err)
	}

	// 无法创建分片会话
	{
		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"POST",
			testMock.Anything,
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 400,
				Body:       ioutil.NopCloser(strings.NewReader(`{"uploadUrl":"123321"}`)),
			},
		})
		client.Request = clientMock
		err := client.Upload(context.Background(), "123.jpg", 15*1024*1024, strings.NewReader("123"))
		clientMock.AssertExpectations(t)
		asserts.Error(err)
	}

}

func TestClient_SimpleUpload(t *testing.T) {
	asserts := assert.New(t)
	client, _ := NewClient(&model.Policy{})
	client.Credential.AccessToken = "AccessToken"
	client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
	cache.Set("setting_onedrive_chunk_retries", "1", 0)

	// 请求失败，并重试
	{
		client.Credential.ExpiresIn = 0
		res, err := client.SimpleUpload(context.Background(), "123.jpg", strings.NewReader("123"), 3)
		asserts.Error(err)
		asserts.Nil(res)
	}

	cache.Set("setting_onedrive_chunk_retries", "0", 0)
	// 返回未知响应
	{
		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"PUT",
			testMock.Anything,
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`???`)),
			},
		})
		client.Request = clientMock
		res, err := client.SimpleUpload(context.Background(), "123.jpg", strings.NewReader("123"), 3)
		clientMock.AssertExpectations(t)
		asserts.Error(err)
		asserts.Nil(res)
	}

	// 返回正常
	{
		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"PUT",
			testMock.Anything,
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`{"name":"123321"}`)),
			},
		})
		client.Request = clientMock
		res, err := client.SimpleUpload(context.Background(), "123.jpg", strings.NewReader("123"), 3)
		clientMock.AssertExpectations(t)
		asserts.NoError(err)
		asserts.NotNil(res)
		asserts.Equal("123321", res.Name)
	}
}

func TestClient_DeleteUploadSession(t *testing.T) {
	asserts := assert.New(t)
	client, _ := NewClient(&model.Policy{})
	client.Credential.AccessToken = "AccessToken"

	// 请求失败
	{
		client.Credential.ExpiresIn = 0
		err := client.DeleteUploadSession(context.Background(), "123.jpg")
		asserts.Error(err)

	}

	// 返回正常
	{
		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"DELETE",
			testMock.Anything,
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 204,
				Body:       ioutil.NopCloser(strings.NewReader(``)),
			},
		})
		client.Request = clientMock
		err := client.DeleteUploadSession(context.Background(), "123.jpg")
		clientMock.AssertExpectations(t)
		asserts.NoError(err)
	}
}

func TestClient_BatchDelete(t *testing.T) {
	asserts := assert.New(t)
	client, _ := NewClient(&model.Policy{})
	client.Credential.AccessToken = "AccessToken"

	// 小于20个，失败1个
	{
		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
		clientMock := ClientMock{}
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
				Body:       ioutil.NopCloser(strings.NewReader(`{"responses":[{"id":"2","status":400}]}`)),
			},
		})
		client.Request = clientMock
		res, err := client.BatchDelete(context.Background(), []string{"1", "2", "3", "1", "2"})
		clientMock.AssertExpectations(t)
		asserts.Error(err)
		asserts.Equal([]string{"2"}, res)
	}
}

func TestClient_Delete(t *testing.T) {
	asserts := assert.New(t)
	client, _ := NewClient(&model.Policy{})
	client.Credential.AccessToken = "AccessToken"

	// 请求失败
	{
		client.Credential.ExpiresIn = 0
		res, err := client.Delete(context.Background(), []string{"1", "2", "3"})
		asserts.Error(err)
		asserts.Len(res, 3)
	}

	// 返回未知响应
	{
		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
		clientMock := ClientMock{}
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
				Body:       ioutil.NopCloser(strings.NewReader(`???`)),
			},
		})
		client.Request = clientMock
		res, err := client.Delete(context.Background(), []string{"1", "2", "3"})
		clientMock.AssertExpectations(t)
		asserts.Error(err)
		asserts.Len(res, 3)
	}

	// 成功2两个文件
	{
		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
		clientMock := ClientMock{}
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
				Body:       ioutil.NopCloser(strings.NewReader(`{"responses":[{"id":"2","status":400}]}`)),
			},
		})
		client.Request = clientMock
		res, err := client.Delete(context.Background(), []string{"1", "2", "3"})
		clientMock.AssertExpectations(t)
		asserts.Error(err)
		asserts.Equal([]string{"2"}, res)
	}
}

func TestClient_ListChildren(t *testing.T) {
	asserts := assert.New(t)
	client, _ := NewClient(&model.Policy{})
	client.Credential.AccessToken = "AccessToken"

	// 根目录，请求失败,重测试
	{
		client.Credential.ExpiresIn = 0
		res, err := client.ListChildren(context.Background(), "/")
		asserts.Error(err)
		asserts.Empty(res)
	}

	// 非根目录，未知响应
	{
		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
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
				Body:       ioutil.NopCloser(strings.NewReader(`???`)),
			},
		})
		client.Request = clientMock
		res, err := client.ListChildren(context.Background(), "/uploads")
		asserts.Error(err)
		asserts.Empty(res)
	}

	// 非根目录，成功
	{
		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
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
				Body:       ioutil.NopCloser(strings.NewReader(`{"value":[{}]}`)),
			},
		})
		client.Request = clientMock
		res, err := client.ListChildren(context.Background(), "/uploads")
		asserts.NoError(err)
		asserts.Len(res, 1)
	}
}

func TestClient_GetThumbURL(t *testing.T) {
	asserts := assert.New(t)
	client, _ := NewClient(&model.Policy{})
	client.Credential.AccessToken = "AccessToken"

	// 请求失败
	{
		client.Credential.ExpiresIn = 0
		res, err := client.GetThumbURL(context.Background(), "123,jpg", 1, 1)
		asserts.Error(err)
		asserts.Empty(res)
	}

	// 未知响应
	{
		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
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
				Body:       ioutil.NopCloser(strings.NewReader(`???`)),
			},
		})
		client.Request = clientMock
		res, err := client.GetThumbURL(context.Background(), "123,jpg", 1, 1)
		asserts.Error(err)
		asserts.Empty(res)
	}

	// 世纪互联 成功
	{
		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
		client.Endpoints.isInChina = true
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
				Body:       ioutil.NopCloser(strings.NewReader(`{"url":"thumb"}`)),
			},
		})
		client.Request = clientMock
		res, err := client.GetThumbURL(context.Background(), "123,jpg", 1, 1)
		asserts.NoError(err)
		asserts.Equal("thumb", res)
	}

	// 非世纪互联 成功
	{
		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
		client.Endpoints.isInChina = false
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
				Body:       ioutil.NopCloser(strings.NewReader(`{"value":[{"large":{"url":"thumb"}}]}`)),
			},
		})
		client.Request = clientMock
		res, err := client.GetThumbURL(context.Background(), "123,jpg", 1, 1)
		asserts.NoError(err)
		asserts.Equal("thumb", res)
	}
}

func TestClient_MonitorUpload(t *testing.T) {
	asserts := assert.New(t)
	client, _ := NewClient(&model.Policy{})
	client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()

	// 客户端完成回调
	{
		cache.Set("setting_onedrive_monitor_timeout", "600", 0)
		cache.Set("setting_onedrive_callback_check", "20", 0)
		asserts.NotPanics(func() {
			go func() {
				time.Sleep(time.Duration(1) * time.Second)
				FinishCallback("key")
			}()
			client.MonitorUpload("url", "key", "path", 10, 10)
		})
	}

	// 上传会话到期，仍未完成上传，创建占位符
	{
		cache.Set("setting_onedrive_monitor_timeout", "600", 0)
		cache.Set("setting_onedrive_callback_check", "20", 0)
		asserts.NotPanics(func() {
			client.MonitorUpload("url", "key", "path", 10, 0)
		})
	}

	fmt.Println("测试:上传已完成，未发送回调")
	// 上传已完成，未发送回调
	{
		cache.Set("setting_onedrive_monitor_timeout", "0", 0)
		cache.Set("setting_onedrive_callback_check", "0", 0)

		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
		client.Credential.AccessToken = "1"
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
				StatusCode: 404,
				Body:       ioutil.NopCloser(strings.NewReader(`{"error":{"code":"itemNotFound"}}`)),
			},
		})
		clientMock.On(
			"Request",
			"POST",
			testMock.Anything,
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 404,
				Body:       ioutil.NopCloser(strings.NewReader(`{"error":{"code":"itemNotFound"}}`)),
			},
		})
		client.Request = clientMock
		cache.Set("callback_key3", "ok", 0)

		asserts.NotPanics(func() {
			client.MonitorUpload("url", "key3", "path", 10, 10)
		})

		clientMock.AssertExpectations(t)
	}

	fmt.Println("测试:上传仍未开始")
	// 上传仍未开始
	{
		cache.Set("setting_onedrive_monitor_timeout", "0", 0)
		cache.Set("setting_onedrive_callback_check", "0", 0)

		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
		client.Credential.AccessToken = "1"
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
				Body:       ioutil.NopCloser(strings.NewReader(`{"nextExpectedRanges":["0-"]}`)),
			},
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
				Body:       ioutil.NopCloser(strings.NewReader(``)),
			},
		})
		clientMock.On(
			"Request",
			"PUT",
			testMock.Anything,
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
			},
		})
		client.Request = clientMock

		asserts.NotPanics(func() {
			client.MonitorUpload("url", "key4", "path", 10, 10)
		})

		clientMock.AssertExpectations(t)
	}

}
