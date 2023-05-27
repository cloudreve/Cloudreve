package filesystem

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/response"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type FileHeaderMock struct {
	testMock.Mock
}

func (m FileHeaderMock) Put(ctx context.Context, file fsctx.FileHeader) error {
	args := m.Called(ctx, file)
	return args.Error(0)
}

func (m FileHeaderMock) Token(ctx context.Context, ttl int64, uploadSession *serializer.UploadSession, file fsctx.FileHeader) (*serializer.UploadCredential, error) {
	args := m.Called(ctx, ttl, uploadSession, file)
	return args.Get(0).(*serializer.UploadCredential), args.Error(1)
}

func (m FileHeaderMock) CancelToken(ctx context.Context, uploadSession *serializer.UploadSession) error {
	args := m.Called(ctx, uploadSession)
	return args.Error(0)
}

func (m FileHeaderMock) List(ctx context.Context, path string, recursive bool) ([]response.Object, error) {
	args := m.Called(ctx, path, recursive)
	return args.Get(0).([]response.Object), args.Error(1)
}

func (m FileHeaderMock) Get(ctx context.Context, path string) (response.RSCloser, error) {
	args := m.Called(ctx, path)
	return args.Get(0).(response.RSCloser), args.Error(1)
}

func (m FileHeaderMock) Delete(ctx context.Context, files []string) ([]string, error) {
	args := m.Called(ctx, files)
	return args.Get(0).([]string), args.Error(1)
}

func (m FileHeaderMock) Thumb(ctx context.Context, files *model.File) (*response.ContentResponse, error) {
	args := m.Called(ctx, files)
	return args.Get(0).(*response.ContentResponse), args.Error(1)
}

func (m FileHeaderMock) Source(ctx context.Context, path string, expires int64, isDownload bool, speed int) (string, error) {
	args := m.Called(ctx, path, expires, isDownload, speed)
	return args.Get(0).(string), args.Error(1)
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
		},
		Policy: &model.Policy{
			AutoRename:  false,
			DirNameRule: "{path}",
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/", nil)
	ctx = context.WithValue(ctx, fsctx.GinCtx, c)
	cancel()
	file := &fsctx.FileStream{
		Size:        5,
		VirtualPath: "/",
		Name:        "1.txt",
	}
	err := fs.Upload(ctx, file)
	asserts.NoError(err)

	// 正常，上下文已指定源文件
	testHandler = new(FileHeaderMock)
	testHandler.On("Put", testMock.Anything, testMock.Anything).Return(nil)
	fs = &FileSystem{
		Handler: testHandler,
		User: &model.User{
			Model: gorm.Model{
				ID: 1,
			},
		},
		Policy: &model.Policy{
			AutoRename:  false,
			DirNameRule: "{path}",
		},
	}
	ctx, cancel = context.WithCancel(context.Background())
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/", nil)
	ctx = context.WithValue(ctx, fsctx.GinCtx, c)
	ctx = context.WithValue(ctx, fsctx.FileModelCtx, model.File{SourceName: "123/123.txt"})
	cancel()
	file = &fsctx.FileStream{
		Size:        5,
		VirtualPath: "/",
		Name:        "1.txt",
		File:        ioutil.NopCloser(strings.NewReader("")),
	}
	err = fs.Upload(ctx, file)
	asserts.NoError(err)

	// BeforeUpload 返回错误
	fs.Use("BeforeUpload", func(ctx context.Context, fs *FileSystem, file fsctx.FileHeader) error {
		return errors.New("error")
	})
	err = fs.Upload(ctx, file)
	asserts.Error(err)
	fs.Hooks["BeforeUpload"] = nil
	testHandler.AssertExpectations(t)

	// 上传文件失败
	testHandler2 := new(FileHeaderMock)
	testHandler2.On("Put", testMock.Anything, testMock.Anything).Return(errors.New("error"))
	fs.Handler = testHandler2
	err = fs.Upload(ctx, file)
	asserts.Error(err)
	testHandler2.AssertExpectations(t)

	// AfterUpload失败
	testHandler3 := new(FileHeaderMock)
	testHandler3.On("Put", testMock.Anything, testMock.Anything).Return(nil)
	fs.Handler = testHandler3
	fs.Use("AfterUpload", func(ctx context.Context, fs *FileSystem, file fsctx.FileHeader) error {
		return errors.New("error")
	})
	fs.Use("AfterValidateFailed", func(ctx context.Context, fs *FileSystem, file fsctx.FileHeader) error {
		return errors.New("error")
	})
	err = fs.Upload(ctx, file)
	asserts.Error(err)
	testHandler2.AssertExpectations(t)

}

func TestFileSystem_GetUploadToken(t *testing.T) {
	asserts := assert.New(t)
	fs := FileSystem{
		User:   &model.User{Model: gorm.Model{ID: 1}},
		Policy: &model.Policy{},
	}
	ctx := context.Background()

	// 成功
	{
		cache.SetSettings(map[string]string{
			"upload_session_timeout": "10",
		}, "setting_")
		testHandler := new(FileHeaderMock)
		testHandler.On("Token", testMock.Anything, int64(10), testMock.Anything, testMock.Anything).Return(&serializer.UploadCredential{Credential: "test"}, nil)
		fs.Handler = testHandler
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(1, 1))
		mock.ExpectQuery("SELECT(.+)files(.+)").WillReturnError(errors.New("not found"))
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)files(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("UPDATE(.+)storage(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		res, err := fs.CreateUploadSession(ctx, &fsctx.FileStream{
			Size:        0,
			Name:        "file",
			VirtualPath: "/",
		})
		asserts.NoError(mock.ExpectationsWereMet())
		testHandler.AssertExpectations(t)
		asserts.NoError(err)
		asserts.Equal("test", res.Credential)
	}

	// 无法获取上传凭证
	{
		cache.SetSettings(map[string]string{
			"upload_credential_timeout": "10",
			"upload_session_timeout":    "10",
		}, "setting_")
		testHandler := new(FileHeaderMock)
		testHandler.On("Token", testMock.Anything, int64(10), testMock.Anything, testMock.Anything).Return(&serializer.UploadCredential{}, errors.New("error"))
		fs.Handler = testHandler
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(1, 1))
		mock.ExpectQuery("SELECT(.+)files(.+)").WillReturnError(errors.New("not found"))
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)files(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("UPDATE(.+)storage(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		_, err := fs.CreateUploadSession(ctx, &fsctx.FileStream{
			Size:        0,
			Name:        "file",
			VirtualPath: "/",
		})
		asserts.NoError(mock.ExpectationsWereMet())
		testHandler.AssertExpectations(t)
		asserts.Error(err)
	}
}

func TestFileSystem_UploadFromStream(t *testing.T) {
	asserts := assert.New(t)
	fs := FileSystem{
		User: &model.User{
			Model:  gorm.Model{ID: 1},
			Policy: model.Policy{Type: "mock"},
		},
		Policy: &model.Policy{Type: "mock"},
	}
	ctx := context.Background()

	err := fs.UploadFromStream(ctx, &fsctx.FileStream{
		File: ioutil.NopCloser(strings.NewReader("123")),
	}, true)
	asserts.Error(err)
}

func TestFileSystem_UploadFromPath(t *testing.T) {
	asserts := assert.New(t)
	fs := FileSystem{
		User: &model.User{
			Model:  gorm.Model{ID: 1},
			Policy: model.Policy{Type: "mock"},
		},
		Policy: &model.Policy{Type: "mock"},
	}
	ctx := context.Background()

	// 文件不存在
	{
		err := fs.UploadFromPath(ctx, "test/not_exist", "/", fsctx.Overwrite)
		asserts.Error(err)
	}

	// 文存在,上传失败
	{
		err := fs.UploadFromPath(ctx, "tests/test.zip", "/", fsctx.Overwrite)
		asserts.Error(err)
	}
}
