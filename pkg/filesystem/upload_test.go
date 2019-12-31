package filesystem

import (
	"context"
	"errors"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/cache"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/filesystem/local"
	"github.com/HFO4/cloudreve/pkg/filesystem/response"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type FileHeaderMock struct {
	testMock.Mock
}

func (m FileHeaderMock) Get(ctx context.Context, path string) (response.RSCloser, error) {
	args := m.Called(ctx, path)
	return args.Get(0).(response.RSCloser), args.Error(1)
}

func (m FileHeaderMock) Put(ctx context.Context, file io.ReadCloser, dst string, size uint64) error {
	args := m.Called(ctx, file, dst)
	return args.Error(0)
}

func (m FileHeaderMock) Delete(ctx context.Context, files []string) ([]string, error) {
	args := m.Called(ctx, files)
	return args.Get(0).([]string), args.Error(1)
}

func (m FileHeaderMock) Thumb(ctx context.Context, files string) (*response.ContentResponse, error) {
	args := m.Called(ctx, files)
	return args.Get(0).(*response.ContentResponse), args.Error(1)
}

func (m FileHeaderMock) Source(ctx context.Context, path string, url url.URL, expires int64, isDownload bool, speed int) (string, error) {
	args := m.Called(ctx, path, url, expires, isDownload, speed)
	return args.Get(0).(string), args.Error(1)
}

func (m FileHeaderMock) Token(ctx context.Context, expires int64, key string) (serializer.UploadCredential, error) {
	args := m.Called(ctx, expires, key)
	return args.Get(0).(serializer.UploadCredential), args.Error(1)
}

func TestFileSystem_Upload(t *testing.T) {
	asserts := assert.New(t)

	// 正常
	testHandller := new(FileHeaderMock)
	testHandller.On("Put", testMock.Anything, testMock.Anything, testMock.Anything).Return(nil)
	fs := &FileSystem{
		Handler: testHandller,
		User: &model.User{
			Model: gorm.Model{
				ID: 1,
			},
			Policy: model.Policy{
				AutoRename:  false,
				DirNameRule: "{path}",
			},
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/", nil)
	ctx = context.WithValue(ctx, fsctx.GinCtx, c)
	cancel()
	file := local.FileStream{
		Size:        5,
		VirtualPath: "/",
		Name:        "1.txt",
	}
	err := fs.Upload(ctx, file)
	asserts.NoError(err)

	// 正常，上下文已指定源文件
	testHandller = new(FileHeaderMock)
	testHandller.On("Put", testMock.Anything, testMock.Anything, "123/123.txt").Return(nil)
	fs = &FileSystem{
		Handler: testHandller,
		User: &model.User{
			Model: gorm.Model{
				ID: 1,
			},
			Policy: model.Policy{
				AutoRename:  false,
				DirNameRule: "{path}",
			},
		},
	}
	ctx, cancel = context.WithCancel(context.Background())
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/", nil)
	ctx = context.WithValue(ctx, fsctx.GinCtx, c)
	ctx = context.WithValue(ctx, fsctx.FileModelCtx, model.File{SourceName: "123/123.txt"})
	cancel()
	file = local.FileStream{
		Size:        5,
		VirtualPath: "/",
		Name:        "1.txt",
	}
	err = fs.Upload(ctx, file)
	asserts.NoError(err)

	// BeforeUpload 返回错误
	fs.Use("BeforeUpload", func(ctx context.Context, fs *FileSystem) error {
		return errors.New("error")
	})
	err = fs.Upload(ctx, file)
	asserts.Error(err)
	fs.Hooks["BeforeUpload"] = nil
	testHandller.AssertExpectations(t)

	// 上传文件失败
	testHandller2 := new(FileHeaderMock)
	testHandller2.On("Put", testMock.Anything, testMock.Anything, testMock.Anything).Return(errors.New("error"))
	fs.Handler = testHandller2
	err = fs.Upload(ctx, file)
	asserts.Error(err)
	testHandller2.AssertExpectations(t)

	// AfterUpload失败
	testHandller3 := new(FileHeaderMock)
	testHandller3.On("Put", testMock.Anything, testMock.Anything, testMock.Anything).Return(nil)
	fs.Handler = testHandller3
	fs.Use("AfterUpload", func(ctx context.Context, fs *FileSystem) error {
		return errors.New("error")
	})
	fs.Use("AfterValidateFailed", func(ctx context.Context, fs *FileSystem) error {
		return errors.New("error")
	})
	err = fs.Upload(ctx, file)
	asserts.Error(err)
	testHandller2.AssertExpectations(t)

}

func TestFileSystem_GenerateSavePath_Anonymous(t *testing.T) {
	asserts := assert.New(t)
	fs := FileSystem{User: &model.User{}}
	ctx := context.WithValue(
		context.Background(),
		fsctx.UploadPolicyCtx,
		serializer.UploadPolicy{
			SavePath:   "{randomkey16}",
			AutoRename: false,
		},
	)

	savePath := fs.GenerateSavePath(ctx, local.FileStream{
		Name: "test.test",
	})
	asserts.Len(savePath, 26)
	asserts.Contains(savePath, "test.test")
}

func TestFileSystem_GetUploadToken(t *testing.T) {
	asserts := assert.New(t)
	fs := FileSystem{User: &model.User{Model: gorm.Model{ID: 1}}}
	ctx := context.Background()

	// 成功
	{
		cache.SetSettings(map[string]string{
			"upload_credential_timeout": "10",
			"upload_session_timeout":    "10",
		}, "setting_")
		testHandller := new(FileHeaderMock)
		testHandller.On("Token", testMock.Anything, int64(10), testMock.Anything).Return(serializer.UploadCredential{Token: "test"}, nil)
		fs.Handler = testHandller
		res, err := fs.GetUploadToken(ctx, "/", 10)
		testHandller.AssertExpectations(t)
		asserts.NoError(err)
		asserts.Equal("test", res.Token)
	}

	// 无法获取上传凭证
	{
		cache.SetSettings(map[string]string{
			"upload_credential_timeout": "10",
			"upload_session_timeout":    "10",
		}, "setting_")
		testHandller := new(FileHeaderMock)
		testHandller.On("Token", testMock.Anything, int64(10), testMock.Anything).Return(serializer.UploadCredential{}, errors.New("error"))
		fs.Handler = testHandller
		_, err := fs.GetUploadToken(ctx, "/", 10)
		testHandller.AssertExpectations(t)
		asserts.Error(err)
	}

	// 上传有效期错误
	{
		cache.SetSettings(map[string]string{
			"upload_credential_timeout": "？10",
			"upload_session_timeout":    "10",
		}, "setting_")

		_, err := fs.GetUploadToken(ctx, "/", 10)
		asserts.Error(err)
	}

	// 上传会话有效期错误
	{
		cache.SetSettings(map[string]string{
			"upload_credential_timeout": "10",
			"upload_session_timeout":    "？10",
		}, "setting_")

		_, err := fs.GetUploadToken(ctx, "/", 10)
		asserts.Error(err)
	}
}
