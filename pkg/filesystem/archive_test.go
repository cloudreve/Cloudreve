package filesystem

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
)

func TestFileSystem_Compress(t *testing.T) {
	asserts := assert.New(t)
	ctx := context.Background()
	fs := FileSystem{
		User: &model.User{Model: gorm.Model{ID: 1}},
	}

	// 成功
	{
		// 查找压缩父目录
		mock.ExpectQuery("SELECT(.+)folders(.+)").
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "parent"))
		// 查找顶级待压缩文件
		mock.ExpectQuery("SELECT(.+)files(.+)").
			WithArgs(1, 1).
			WillReturnRows(
				sqlmock.NewRows(
					[]string{"id", "name", "source_name", "policy_id"}).
					AddRow(1, "1.txt", "tests/file1.txt", 1),
			)
		asserts.NoError(cache.Set("setting_temp_path", "tests", -1))
		// 查找父目录子文件
		mock.ExpectQuery("SELECT(.+)files(.+)").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "source_name", "policy_id"}))
		// 查找子目录
		mock.ExpectQuery("SELECT(.+)folders(.+)").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(2, "sub"))
		// 查找子目录子文件
		mock.ExpectQuery("SELECT(.+)files(.+)").
			WithArgs(2).
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "name", "source_name", "policy_id"}).
					AddRow(2, "2.txt", "tests/file2.txt", 1),
			)
		// 查找上传策略
		asserts.NoError(cache.Set("policy_1", model.Policy{Type: "local"}, -1))

		zipFile, err := fs.Compress(ctx, []uint{1}, []uint{1}, true)
		asserts.NoError(err)
		asserts.NotEmpty(zipFile)
		asserts.Contains(zipFile, "archive_")
		asserts.Contains(zipFile, "tests")
	}

	// 上下文取消
	{
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		// 查找压缩父目录
		mock.ExpectQuery("SELECT(.+)folders(.+)").
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "parent"))
		// 查找顶级待压缩文件
		mock.ExpectQuery("SELECT(.+)files(.+)").
			WithArgs(1, 1).
			WillReturnRows(
				sqlmock.NewRows(
					[]string{"id", "name", "source_name", "policy_id"}).
					AddRow(1, "1.txt", "tests/file1.txt", 1),
			)
		asserts.NoError(cache.Set("setting_temp_path", "tests", -1))

		zipFile, err := fs.Compress(ctx, []uint{1}, []uint{1}, true)
		asserts.Error(err)
		asserts.Empty(zipFile)
	}

	// 限制父目录
	{
		ctx := context.WithValue(context.Background(), fsctx.LimitParentCtx, &model.Folder{
			Model: gorm.Model{ID: 3},
		})
		// 查找压缩父目录
		mock.ExpectQuery("SELECT(.+)folders(.+)").
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "parent_id"}).AddRow(1, "parent", 3))
		// 查找顶级待压缩文件
		mock.ExpectQuery("SELECT(.+)files(.+)").
			WithArgs(1, 1).
			WillReturnRows(
				sqlmock.NewRows(
					[]string{"id", "name", "source_name", "policy_id"}).
					AddRow(1, "1.txt", "tests/file1.txt", 1),
			)
		asserts.NoError(cache.Set("setting_temp_path", "tests", -1))

		zipFile, err := fs.Compress(ctx, []uint{1}, []uint{1}, true)
		asserts.Error(err)
		asserts.Equal(ErrObjectNotExist, err)
		asserts.Empty(zipFile)
	}

}

type MockNopRSC string

func (m MockNopRSC) Read(b []byte) (int, error) {
	return 0, errors.New("read error")
}

func (m MockNopRSC) Seek(n int64, offset int) (int64, error) {
	return 0, errors.New("read error")
}

func (m MockNopRSC) Close() error {
	return errors.New("read error")
}

type MockRSC struct {
	rs io.ReadSeeker
}

func (m MockRSC) Read(b []byte) (int, error) {
	return m.rs.Read(b)
}

func (m MockRSC) Seek(n int64, offset int) (int64, error) {
	return m.rs.Seek(n, offset)
}

func (m MockRSC) Close() error {
	return nil
}

func TestFileSystem_Decompress(t *testing.T) {
	asserts := assert.New(t)
	ctx := context.Background()
	fs := FileSystem{
		User: &model.User{Model: gorm.Model{ID: 1}},
	}

	// 压缩文件不存在
	{
		// 查找根目录
		mock.ExpectQuery("SELECT(.+)folders(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "/"))
		// 查找压缩文件，未找到
		mock.ExpectQuery("SELECT(.+)files(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
		err := fs.Decompress(ctx, "/1.zip", "/")
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
	}

	// 无法下载压缩文件
	{
		fs.FileTarget = []model.File{{SourceName: "1.zip", Policy: model.Policy{Type: "mock"}}}
		fs.FileTarget[0].Policy.ID = 1
		testHandler := new(FileHeaderMock)
		testHandler.On("Get", testMock.Anything, "1.zip").Return(request.NopRSCloser{}, errors.New("error"))
		fs.Handler = testHandler
		err := fs.Decompress(ctx, "/1.zip", "/")
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.EqualError(err, "error")
	}

	// 无法创建临时压缩文件
	{
		cache.Set("setting_temp_path", "/tests:", 0)
		fs.FileTarget = []model.File{{SourceName: "1.zip", Policy: model.Policy{Type: "mock"}}}
		fs.FileTarget[0].Policy.ID = 1
		testHandler := new(FileHeaderMock)
		testHandler.On("Get", testMock.Anything, "1.zip").Return(request.NopRSCloser{}, nil)
		fs.Handler = testHandler
		err := fs.Decompress(ctx, "/1.zip", "/")
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
	}

	// 无法写入压缩文件
	{
		cache.Set("setting_temp_path", "tests", 0)
		fs.FileTarget = []model.File{{SourceName: "1.zip", Policy: model.Policy{Type: "mock"}}}
		fs.FileTarget[0].Policy.ID = 1
		testHandler := new(FileHeaderMock)
		testHandler.On("Get", testMock.Anything, "1.zip").Return(MockNopRSC("1"), nil)
		fs.Handler = testHandler
		err := fs.Decompress(ctx, "/1.zip", "/")
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.EqualError(err, "read error")
	}

	// 无效zip文件
	{
		cache.Set("setting_temp_path", "tests", 0)
		fs.FileTarget = []model.File{{SourceName: "1.zip", Policy: model.Policy{Type: "mock"}}}
		fs.FileTarget[0].Policy.ID = 1
		testHandler := new(FileHeaderMock)
		testHandler.On("Get", testMock.Anything, "1.zip").Return(MockRSC{rs: strings.NewReader("read")}, nil)
		fs.Handler = testHandler
		err := fs.Decompress(ctx, "/1.zip", "/")
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.EqualError(err, "zip: not a valid zip file")
	}

	// 无法重设上传策略
	{
		zipFile, _ := os.Open(util.RelativePath("filesystem/tests/test.zip"))
		fs.FileTarget = []model.File{{SourceName: "1.zip", Policy: model.Policy{Type: "mock"}}}
		fs.FileTarget[0].Policy.ID = 1
		testHandler := new(FileHeaderMock)
		testHandler.On("Get", testMock.Anything, "1.zip").Return(zipFile, nil)
		fs.Handler = testHandler
		err := fs.Decompress(ctx, "/1.zip", "/")
		zipFile.Close()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.True(util.IsEmpty(util.RelativePath("tests/decompress")))
	}

	// 无法上传，容量不足
	{
		cache.Set("setting_max_parallel_transfer", "1", 0)
		zipFile, _ := os.Open(util.RelativePath("filesystem/tests/test.zip"))
		fs.FileTarget = []model.File{{SourceName: "1.zip", Policy: model.Policy{Type: "mock"}}}
		fs.FileTarget[0].Policy.ID = 1
		fs.User.Policy.Type = "mock"
		testHandler := new(FileHeaderMock)
		testHandler.On("Get", testMock.Anything, "1.zip").Return(zipFile, nil)
		fs.Handler = testHandler

		fs.Decompress(ctx, "/1.zip", "/")

		zipFile.Close()

		asserts.NoError(mock.ExpectationsWereMet())
		testHandler.AssertExpectations(t)
	}
}
