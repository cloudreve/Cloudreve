package onedrive

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/chunk"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/chunk/backoff"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/mq"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
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

	// OneDrive返回429错误
	{
		header := http.Header{}
		header.Add("retry-after", "120")
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
				StatusCode: 429,
				Header:     header,
				Body:       ioutil.NopCloser(strings.NewReader(`{"error":{"message":"error msg"}}`)),
			},
		})
		client.Request = clientMock
		res, err := client.request(context.Background(), "POST", "http://dev.com", strings.NewReader(""))
		clientMock.AssertExpectations(t)
		asserts.Error(err)
		asserts.Empty(res)
		var retryErr *backoff.RetryableError
		asserts.ErrorAs(err, &retryErr)
		asserts.EqualValues(time.Duration(120)*time.Second, retryErr.RetryAfter)
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
				Path: "/drive/root:/123/32%201",
			},
		}
		asserts.Equal("123/32 1/%e6%96%87%e4%bb%b6%e5%90%8d.jpg", fileInfo.GetSourcePath())
	}

	// 失败
	{
		fileInfo := FileInfo{
			Name: "123.jpg",
			ParentReference: parentReference{
				Path: "/drive/root:/123/%e6%96%87%e4%bb%b6%e5%90%8g",
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

	// 返回正常, 使用资源id
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
		res, err := client.Meta(context.Background(), "123321", "123")
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
	cg := chunk.NewChunkGroup(&fsctx.FileStream{Size: 15}, 10, &backoff.ConstantBackoff{}, false)

	// 非最后分片，正常
	{
		cg.Next()
		client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"PUT",
			"http://dev.com",
			testMock.Anything,
			testMock.Anything,
			testMock.Anything,
			testMock.Anything,
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
		res, err := client.UploadChunk(context.Background(), "http://dev.com", strings.NewReader("1234567890"), cg)
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
		res, err := client.UploadChunk(context.Background(), "http://dev.com", strings.NewReader("1234567890"), cg)
		clientMock.AssertExpectations(t)
		asserts.Error(err)
		asserts.Nil(res)
	}

	// 最后分片，正常
	{
		cg.Next()
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
		res, err := client.UploadChunk(context.Background(), "http://dev.com", strings.NewReader("12345"), cg)
		clientMock.AssertExpectations(t)
		asserts.NoError(err)
		asserts.Nil(res)
	}

	// 最后分片，失败
	{
		cache.Set("setting_chunk_retries", "1", 0)
		client.Credential.ExpiresIn = 0
		go func() {
			time.Sleep(time.Duration(2) * time.Second)
			client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
		}()
		clientMock := ClientMock{}
		client.Request = clientMock
		res, err := client.UploadChunk(context.Background(), "http://dev.com", strings.NewReader("12345"), cg)
		clientMock.AssertExpectations(t)
		asserts.Error(err)
		asserts.Nil(res)
	}
}

func TestClient_Upload(t *testing.T) {
	asserts := assert.New(t)
	client, _ := NewClient(&model.Policy{})
	client.Credential.AccessToken = "AccessToken"
	client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
	ctx := context.Background()
	cache.Set("setting_chunk_retries", "1", 0)
	cache.Set("setting_use_temp_chunk_buffer", "false", 0)

	// 小文件，简单上传，失败
	{
		client.Credential.ExpiresIn = 0
		err := client.Upload(ctx, &fsctx.FileStream{
			Size: 5,
			File: io.NopCloser(strings.NewReader("12345")),
		})
		asserts.Error(err)
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
		err := client.Upload(context.Background(), &fsctx.FileStream{
			Size: SmallFileSize + 1,
			File: io.NopCloser(strings.NewReader("12345")),
		})
		clientMock.AssertExpectations(t)
		asserts.Error(err)
	}

	// 分片上传失败
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
		clientMock.On(
			"Request",
			"PUT",
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
		err := client.Upload(context.Background(), &fsctx.FileStream{
			Size: SmallFileSize + 1,
			File: io.NopCloser(strings.NewReader("12345")),
		})
		clientMock.AssertExpectations(t)
		asserts.Error(err)
		asserts.Contains(err.Error(), "failed to upload chunk")
	}

}

func TestClient_SimpleUpload(t *testing.T) {
	asserts := assert.New(t)
	client, _ := NewClient(&model.Policy{})
	client.Credential.AccessToken = "AccessToken"
	client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
	cache.Set("setting_chunk_retries", "1", 0)

	// 请求失败
	{
		client.Credential.ExpiresIn = 0
		res, err := client.SimpleUpload(context.Background(), "123.jpg", strings.NewReader("123"), 3)
		asserts.Error(err)
		asserts.Nil(res)
	}

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
				mq.GlobalMQ.Publish("key", mq.Message{})
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
