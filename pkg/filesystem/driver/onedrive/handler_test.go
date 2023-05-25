package onedrive

import (
	"context"
	"fmt"
	"github.com/cloudreve/Cloudreve/v3/pkg/mq"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/jinzhu/gorm"
	"io"
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

func TestDriver_Token(t *testing.T) {
	asserts := assert.New(t)
	h, _ := NewDriver(&model.Policy{
		AccessKey:  "ak",
		SecretKey:  "sk",
		BucketName: "test",
		Server:     "test.com",
	})
	handler := h.(Driver)

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
		res, err := handler.Token(context.Background(), 10, &serializer.UploadSession{}, &fsctx.FileStream{})
		asserts.Error(err)
		asserts.Nil(res)
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
		go func() {
			time.Sleep(time.Duration(1) * time.Second)
			mq.GlobalMQ.Publish("TestDriver_Token", mq.Message{})
		}()
		res, err := handler.Token(context.Background(), 10, &serializer.UploadSession{Key: "TestDriver_Token"}, &fsctx.FileStream{})
		asserts.NoError(err)
		asserts.Equal("123321", res.UploadURLs[0])
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
		res, err := handler.Source(context.Background(), "123.jpg", 1, true, 0)
		asserts.Error(err)
		asserts.Empty(res)
	}

	// 命中缓存 成功
	{
		handler.Client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
		handler.Client.Credential.AccessToken = "1"
		cache.Set("onedrive_source_0_123.jpg", "res", 1)
		res, err := handler.Source(context.Background(), "123.jpg", 0, true, 0)
		cache.Deletes([]string{"0_123.jpg"}, "onedrive_source_")
		asserts.NoError(err)
		asserts.Equal("res", res)
	}

	// 命中缓存 上下文存在文件 成功
	{
		file := model.File{}
		file.ID = 1
		file.UpdatedAt = time.Now()
		ctx := context.WithValue(context.Background(), fsctx.FileModelCtx, file)
		handler.Client.Credential.ExpiresIn = time.Now().Add(time.Duration(100) * time.Hour).Unix()
		handler.Client.Credential.AccessToken = "1"
		cache.Set(fmt.Sprintf("onedrive_source_file_%d_1", file.UpdatedAt.Unix()), "res", 0)
		res, err := handler.Source(ctx, "123.jpg", 1, true, 0)
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
		res, err := handler.Source(context.Background(), "123.jpg", 1, true, 0)
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
	file := &model.File{PicInfo: "1,1", Model: gorm.Model{ID: 1}}

	// 失败
	{
		ctx := context.WithValue(context.Background(), fsctx.ThumbSizeCtx, [2]uint{10, 20})
		res, err := handler.Thumb(ctx, file)
		asserts.Error(err)
		asserts.Empty(res.URL)
	}

	// 上下文错误
	{
		_, err := handler.Thumb(context.Background(), file)
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
		err := handler.Put(context.Background(), &fsctx.FileStream{})
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

func TestDriver_replaceSourceHost(t *testing.T) {
	tests := []struct {
		name    string
		origin  string
		cdn     string
		want    string
		wantErr bool
	}{
		{"TestNoReplace", "http://1dr.ms/download.aspx?123456", "", "http://1dr.ms/download.aspx?123456", false},
		{"TestReplaceCorrect", "http://1dr.ms/download.aspx?123456", "https://test.com:8080", "https://test.com:8080/download.aspx?123456", false},
		{"TestCdnFormatError", "http://1dr.ms/download.aspx?123456", string([]byte{0x7f}), "", true},
		{"TestSrcFormatError", string([]byte{0x7f}), "https://test.com:8080", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := &model.Policy{}
			policy.OptionsSerialized.OdProxy = tt.cdn
			handler := Driver{
				Policy: policy,
			}
			got, err := handler.replaceSourceHost(tt.origin)
			if (err != nil) != tt.wantErr {
				t.Errorf("replaceSourceHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("replaceSourceHost() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDriver_CancelToken(t *testing.T) {
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
		err := handler.CancelToken(context.Background(), &serializer.UploadSession{})
		asserts.Error(err)
	}
}
