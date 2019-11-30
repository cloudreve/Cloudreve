package filesystem

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem/local"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestFileSystem_AddFile(t *testing.T) {
	asserts := assert.New(t)
	file := local.FileStream{
		Size: 5,
		Name: "1.txt",
	}
	folder := model.Folder{
		Model: gorm.Model{
			ID: 1,
		},
		PositionAbsolute: "/我的文件",
	}
	fs := FileSystem{
		User: &model.User{
			Model: gorm.Model{
				ID: 1,
			},
			Policy: model.Policy{
				Model: gorm.Model{
					ID: 1,
				},
			},
		},
	}
	ctx := context.WithValue(context.Background(), FileHeaderCtx, file)
	ctx = context.WithValue(ctx, SavePathCtx, "/Uploads/1_sad.txt")

	_, err := fs.AddFile(ctx, &folder)

	asserts.Error(err)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	f, err := fs.AddFile(ctx, &folder)

	asserts.NoError(err)
	asserts.Equal("/Uploads/1_sad.txt", f.SourceName)
}

func TestFileSystem_GetContent(t *testing.T) {
	asserts := assert.New(t)
	ctx := context.Background()
	fs := FileSystem{
		User: &model.User{
			Model: gorm.Model{
				ID: 1,
			},
			Policy: model.Policy{
				Model: gorm.Model{
					ID: 1,
				},
			},
		},
	}

	// 文件不存在
	rs, err := fs.GetContent(ctx, "not exist file")
	asserts.Equal(ErrObjectNotExist, err)
	asserts.Nil(rs)

	// 未知存储策略
	file, err := os.Create("TestFileSystem_GetContent.txt")
	asserts.NoError(err)
	_ = file.Close()

	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "policy_id"}).AddRow(1, "TestFileSystem_GetContent.txt", 1))
	mock.ExpectQuery("SELECT(.+)poli(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(1, "unknown"))

	rs, err = fs.GetContent(ctx, "TestFileSystem_GetContent.txt")
	asserts.Error(err)
	asserts.NoError(mock.ExpectationsWereMet())

	// 打开文件失败
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "policy_id"}).AddRow(1, "TestFileSystem_GetContent.txt", 1))
	mock.ExpectQuery("SELECT(.+)poli(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "type", "source_name"}).AddRow(1, "local", "not exist"))

	rs, err = fs.GetContent(ctx, "TestFileSystem_GetContent.txt")
	asserts.Equal(serializer.CodeIOFailed, err.(serializer.AppError).Code)
	asserts.NoError(mock.ExpectationsWereMet())

	// 打开成功
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "policy_id", "source_name"}).AddRow(1, "TestFileSystem_GetContent.txt", 1, "TestFileSystem_GetContent.txt"))
	mock.ExpectQuery("SELECT(.+)poli(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(1, "local"))

	rs, err = fs.GetContent(ctx, "TestFileSystem_GetContent.txt")
	asserts.NoError(err)
	asserts.NoError(mock.ExpectationsWereMet())
}

func TestFileSystem_GetDownloadContent(t *testing.T) {
	asserts := assert.New(t)
	ctx := context.Background()
	fs := FileSystem{
		User: &model.User{
			Model: gorm.Model{
				ID: 1,
			},
			Policy: model.Policy{
				Model: gorm.Model{
					ID: 1,
				},
			},
		},
	}
	file, err := os.Create("TestFileSystem_GetDownloadContent.txt")
	asserts.NoError(err)
	_ = file.Close()

	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "policy_id", "source_name"}).AddRow(1, "TestFileSystem_GetDownloadContent.txt", 1, "TestFileSystem_GetDownloadContent.txt"))
	mock.ExpectQuery("SELECT(.+)poli(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(1, "local"))

	// 无限速
	_, err = fs.GetDownloadContent(ctx, "TestFileSystem_GetDownloadContent.txt")
	asserts.NoError(err)
	asserts.NoError(mock.ExpectationsWereMet())

	// 有限速
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "policy_id", "source_name"}).AddRow(1, "TestFileSystem_GetDownloadContent.txt", 1, "TestFileSystem_GetDownloadContent.txt"))
	mock.ExpectQuery("SELECT(.+)poli(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(1, "local"))

	fs.User.Group.SpeedLimit = 1
	_, err = fs.GetDownloadContent(ctx, "TestFileSystem_GetDownloadContent.txt")
	asserts.NoError(err)
	asserts.NoError(mock.ExpectationsWereMet())
}
