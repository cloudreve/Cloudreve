package filesystem

import (
	"context"
	"errors"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/filesystem/local"
	"github.com/HFO4/cloudreve/pkg/filesystem/response"
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

func (m FileHeaderMock) Get(ctx context.Context, path string) (io.ReadSeeker, error) {
	args := m.Called(ctx, path)
	return args.Get(0).(io.ReadSeeker), args.Error(1)
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

func (m FileHeaderMock) Source(ctx context.Context, path string, url url.URL, expires int64) (string, error) {
	args := m.Called(ctx, path, url, expires)
	return args.Get(0).(string), args.Error(1)
}
func (m FileHeaderMock) GetDownloadURL(ctx context.Context, path string, url url.URL, expires int64) (string, error) {
	args := m.Called(ctx, path, url, expires)
	return args.Get(0).(string), args.Error(1)
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

	// BeforeUpload 返回错误
	fs.Use("BeforeUpload", func(ctx context.Context, fs *FileSystem) error {
		return errors.New("error")
	})
	err = fs.Upload(ctx, file)
	asserts.Error(err)
	fs.BeforeUpload = nil
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
