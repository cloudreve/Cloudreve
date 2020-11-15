package filesystem

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/local"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/response"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
)

type FileHeaderMock struct {
	testMock.Mock
}

func (m FileHeaderMock) List(ctx context.Context, path string, recursive bool) ([]response.Object, error) {
	args := m.Called(ctx, path, recursive)
	return args.Get(0).([]response.Object), args.Error(1)
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
	testHandler := new(FileHeaderMock)
	testHandler.On("Put", testMock.Anything, testMock.Anything, testMock.Anything).Return(nil)
	fs := &FileSystem{
		Handler: testHandler,
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
	testHandler = new(FileHeaderMock)
	testHandler.On("Put", testMock.Anything, testMock.Anything, "123/123.txt").Return(nil)
	fs = &FileSystem{
		Handler: testHandler,
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
		File:        ioutil.NopCloser(strings.NewReader("")),
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
	testHandler.AssertExpectations(t)

	// 上传文件失败
	testHandler2 := new(FileHeaderMock)
	testHandler2.On("Put", testMock.Anything, testMock.Anything, testMock.Anything).Return(errors.New("error"))
	fs.Handler = testHandler2
	err = fs.Upload(ctx, file)
	asserts.Error(err)
	testHandler2.AssertExpectations(t)

	// AfterUpload失败
	testHandler3 := new(FileHeaderMock)
	testHandler3.On("Put", testMock.Anything, testMock.Anything, testMock.Anything).Return(nil)
	fs.Handler = testHandler3
	fs.Use("AfterUpload", func(ctx context.Context, fs *FileSystem) error {
		return errors.New("error")
	})
	fs.Use("AfterValidateFailed", func(ctx context.Context, fs *FileSystem) error {
		return errors.New("error")
	})
	err = fs.Upload(ctx, file)
	asserts.Error(err)
	testHandler2.AssertExpectations(t)

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
		testHandler := new(FileHeaderMock)
		testHandler.On("Token", testMock.Anything, int64(10), testMock.Anything).Return(serializer.UploadCredential{Token: "test"}, nil)
		fs.Handler = testHandler
		res, err := fs.GetUploadToken(ctx, "/", 10, "123")
		testHandler.AssertExpectations(t)
		asserts.NoError(err)
		asserts.Equal("test", res.Token)
	}

	// 无法获取上传凭证
	{
		cache.SetSettings(map[string]string{
			"upload_credential_timeout": "10",
			"upload_session_timeout":    "10",
		}, "setting_")
		testHandler := new(FileHeaderMock)
		testHandler.On("Token", testMock.Anything, int64(10), testMock.Anything).Return(serializer.UploadCredential{}, errors.New("error"))
		fs.Handler = testHandler
		_, err := fs.GetUploadToken(ctx, "/", 10, "123")
		testHandler.AssertExpectations(t)
		asserts.Error(err)
	}
}

func TestFileSystem_UploadFromStream(t *testing.T) {
	asserts := assert.New(t)
	fs := FileSystem{User: &model.User{Model: gorm.Model{ID: 1}}}
	ctx := context.Background()

	err := fs.UploadFromStream(ctx, ioutil.NopCloser(strings.NewReader("123")), "/1.txt", 1)
	asserts.Error(err)
}

func TestFileSystem_UploadFromPath(t *testing.T) {
	asserts := assert.New(t)
	fs := FileSystem{User: &model.User{Policy: model.Policy{Type: "mock"}, Model: gorm.Model{ID: 1}}}
	ctx := context.Background()

	// 文件不存在
	{
		err := fs.UploadFromPath(ctx, "test/not_exist", "/")
		asserts.Error(err)
	}

	// 文存在,上传失败
	{
		err := fs.UploadFromPath(ctx, "tests/test.zip", "/")
		asserts.Error(err)
	}
}
