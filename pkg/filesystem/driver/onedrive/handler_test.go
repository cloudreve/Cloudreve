package onedrive

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
)

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

	// 无法获取文件路径
	{
		ctx := context.WithValue(context.Background(), fsctx.FileSizeCtx, uint64(10))
		res, err := handler.Token(ctx, 10, "key")
		asserts.Error(err)
		asserts.Equal(serializer.UploadCredential{}, res)
	}

	// 无法获取文件大小
	{
		ctx := context.WithValue(context.Background(), fsctx.SavePathCtx, "/123")
		res, err := handler.Token(ctx, 10, "key")
		asserts.Error(err)
		asserts.Equal(serializer.UploadCredential{}, res)
	}

	// 小文件成功
	{
		ctx := context.WithValue(context.Background(), fsctx.SavePathCtx, "/123")
		ctx = context.WithValue(ctx, fsctx.FileSizeCtx, uint64(10))
		res, err := handler.Token(ctx, 10, "key")
		asserts.NoError(err)
		asserts.Equal(serializer.UploadCredential{}, res)
	}

	// 分片上传 失败
	{
		cache.Set("setting_siteURL", "http://test.cloudreve.org", 0)
		handler.Client, _ = NewClient(&model.Policy{})
		handler.Client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
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
		handler.Client.Request = clientMock
		ctx := context.WithValue(context.Background(), fsctx.SavePathCtx, "/123")
		ctx = context.WithValue(ctx, fsctx.FileSizeCtx, uint64(20*1024*1024))
		res, err := handler.Token(ctx, 10, "key")
		asserts.Error(err)
		asserts.Equal(serializer.UploadCredential{}, res)
	}

	// 分片上传 成功
	{
		cache.Set("setting_siteURL", "http://test.cloudreve.org", 0)
		cache.Set("setting_onedrive_monitor_timeout", "600", 0)
		cache.Set("setting_onedrive_callback_check", "20", 0)
		handler.Client, _ = NewClient(&model.Policy{})
		handler.Client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
		handler.Client.Credential.AccessToken = "1"
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
		handler.Client.Request = clientMock
		ctx := context.WithValue(context.Background(), fsctx.SavePathCtx, "/123")
		ctx = context.WithValue(ctx, fsctx.FileSizeCtx, uint64(20*1024*1024))
		go func() {
			time.Sleep(time.Duration(1) * time.Second)
			FinishCallback("key")
		}()
		res, err := handler.Token(ctx, 10, "key")
		asserts.NoError(err)
		asserts.Equal("123321", res.Policy)
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
		},
	}
	handler.Client, _ = NewClient(&model.Policy{})
	handler.Client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
	cache.Set("setting_onedrive_source_timeout", "1800", 0)

	// 失败
	{
		res, err := handler.Source(context.Background(), "123.jpg", url.URL{}, 0, true, 0)
		asserts.Error(err)
		asserts.Empty(res)
	}

	// 命中缓存 成功
	{
		handler.Client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
		handler.Client.Credential.AccessToken = "1"
		cache.Set("onedrive_source_0_123.jpg", "res", 0)
		res, err := handler.Source(context.Background(), "123.jpg", url.URL{}, 0, true, 0)
		cache.Deletes([]string{"0_123.jpg"}, "onedrive_source_")
		asserts.NoError(err)
		asserts.Equal("res", res)
	}

	// 成功
	{
		handler.Client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
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
				Body:       ioutil.NopCloser(strings.NewReader(`{"@microsoft.graph.downloadUrl":"123321"}`)),
			},
		})
		handler.Client.Request = clientMock
		handler.Client.Credential.AccessToken = "1"
		res, err := handler.Source(context.Background(), "123.jpg", url.URL{}, 0, true, 0)
		asserts.NoError(err)
		asserts.Equal("123321", res)
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
		},
	}
	handler.Client, _ = NewClient(&model.Policy{})
	handler.Client.Credential.AccessToken = "AccessToken"
	handler.Client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()

	// 非递归
	{
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
		handler.Client.Request = clientMock
		res, err := handler.List(context.Background(), "/", false)
		asserts.NoError(err)
		asserts.Len(res, 1)
	}

	// 递归一次
	{
		clientMock := ClientMock{}
		clientMock.On(
			"Request",
			"GET",
			"me/drive/root/children?$top=999999999",
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`{"value":[{"name":"1","folder":{}}]}`)),
			},
		})
		clientMock.On(
			"Request",
			"GET",
			"me/drive/root:/1:/children?$top=999999999",
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`{"value":[{"name":"2"}]}`)),
			},
		})
		handler.Client.Request = clientMock
		res, err := handler.List(context.Background(), "/", true)
		asserts.NoError(err)
		asserts.Len(res, 2)
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
	handler.Client, _ = NewClient(&model.Policy{})
	handler.Client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()

	// 失败
	{
		ctx := context.WithValue(context.Background(), fsctx.ThumbSizeCtx, [2]uint{10, 20})
		ctx = context.WithValue(ctx, fsctx.FileModelCtx, model.File{})
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		res, err := handler.Thumb(ctx, "123.jpg")
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.Empty(res.URL)
	}

	// 上下文错误
	{
		_, err := handler.Thumb(context.Background(), "123.jpg")
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
	}
}

func TestDriver_Delete(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{
		Policy: &model.Policy{
			AccessKey:  "ak",
			SecretKey:  "sk",
			BucketName: "test",
			Server:     "test.com",
		},
	}
	handler.Client, _ = NewClient(&model.Policy{})
	handler.Client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()

	// 失败
	{
		_, err := handler.Delete(context.Background(), []string{"1"})
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
	}

}

func TestDriver_Put(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{
		Policy: &model.Policy{
			AccessKey:  "ak",
			SecretKey:  "sk",
			BucketName: "test",
			Server:     "test.com",
		},
	}
	handler.Client, _ = NewClient(&model.Policy{})
	handler.Client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()

	// 失败
	{
		err := handler.Put(context.Background(), ioutil.NopCloser(strings.NewReader("")), "dst", 0)
		asserts.Error(err)
	}
}

func TestDriver_Get(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{
		Policy: &model.Policy{
			AccessKey:  "ak",
			SecretKey:  "sk",
			BucketName: "test",
			Server:     "test.com",
		},
	}
	handler.Client, _ = NewClient(&model.Policy{})
	handler.Client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()

	// 无法获取source
	{
		res, err := handler.Get(context.Background(), "123.txt")
		asserts.Error(err)
		asserts.Nil(res)
	}

	// 成功
	handler.Client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
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
			Body:       ioutil.NopCloser(strings.NewReader(`{"@microsoft.graph.downloadUrl":"123321"}`)),
		},
	})
	handler.Client.Request = clientMock
	handler.Client.Credential.AccessToken = "1"

	driverClientMock := ClientMock{}
	driverClientMock.On(
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
	handler.HTTPClient = driverClientMock
	res, err := handler.Get(context.Background(), "123.txt")
	clientMock.AssertExpectations(t)
	asserts.NoError(err)
	_, err = res.Seek(0, io.SeekEnd)
	asserts.NoError(err)
	content, err := ioutil.ReadAll(res)
	asserts.NoError(err)
	asserts.Equal("123", string(content))
}
