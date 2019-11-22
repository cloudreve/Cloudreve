package filesystem

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem/local"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGenericBeforeUpload(t *testing.T) {
	asserts := assert.New(t)
	file := local.FileData{
		Size: 5,
		Name: "1.txt",
	}
	ctx := context.WithValue(context.Background(), FileHeaderCtx, file)
	fs := FileSystem{
		User: &model.User{
			Storage: 0,
			Group: model.Group{
				MaxStorage: 11,
			},
			Policy: model.Policy{
				MaxSize: 4,
				OptionsSerialized: model.PolicyOption{
					FileType: []string{"txt"},
				},
			},
		},
	}

	asserts.Error(GenericBeforeUpload(ctx, &fs))

	file.Size = 1
	file.Name = "1"
	ctx = context.WithValue(context.Background(), FileHeaderCtx, file)
	asserts.Error(GenericBeforeUpload(ctx, &fs))

	file.Name = "1.txt"
	ctx = context.WithValue(context.Background(), FileHeaderCtx, file)
	asserts.NoError(GenericBeforeUpload(ctx, &fs))

	file.Name = "1.t/xt"
	ctx = context.WithValue(context.Background(), FileHeaderCtx, file)
	asserts.Error(GenericBeforeUpload(ctx, &fs))
}

func TestGenericAfterUploadCanceled(t *testing.T) {
	asserts := assert.New(t)
	f, err := os.Create("TestGenericAfterUploadCanceled")
	asserts.NoError(err)
	f.Close()
	file := local.FileStream{
		Size: 5,
		Name: "TestGenericAfterUploadCanceled",
	}
	ctx := context.WithValue(context.Background(), SavePathCtx, "TestGenericAfterUploadCanceled")
	ctx = context.WithValue(ctx, FileHeaderCtx, file)
	fs := FileSystem{
		User:    &model.User{Storage: 5},
		Handler: local.Handler{},
	}

	// 成功
	err = GenericAfterUploadCanceled(ctx, &fs)
	asserts.NoError(err)
	asserts.Equal(uint64(0), fs.User.Storage)

	f, err = os.Create("TestGenericAfterUploadCanceled")
	asserts.NoError(err)
	f.Close()

	// 容量不能再降低
	err = GenericAfterUploadCanceled(ctx, &fs)
	asserts.Error(err)

	//文件不存在
	fs.User.Storage = 5
	err = GenericAfterUploadCanceled(ctx, &fs)
	asserts.NoError(err)
}

func TestGenericAfterUpload(t *testing.T) {
	asserts := assert.New(t)
	fs := FileSystem{
		User: &model.User{
			Model: gorm.Model{
				ID: 1,
			},
		},
	}

	ctx := context.WithValue(context.Background(), FileHeaderCtx, local.FileStream{
		VirtualPath: "/我的文件",
		Name:        "test.txt",
	})
	ctx = context.WithValue(ctx, SavePathCtx, "")

	// 正常
	mock.ExpectQuery("SELECT(.+)folders(.+)").WillReturnRows(
		mock.NewRows([]string{"name"}).AddRow("我的文件"),
	)
	mock.ExpectQuery("SELECT(.+)files(.+)").WillReturnError(errors.New("not found"))
	mock.ExpectBegin()
	mock.ExpectExec("INSERT(.+)files(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := GenericAfterUpload(ctx, &fs)
	asserts.NoError(err)
	asserts.NoError(mock.ExpectationsWereMet())

	// 路径不存在
	mock.ExpectQuery("SELECT(.+)folders(.+)").WillReturnRows(
		mock.NewRows([]string{"name"}),
	)
	err = GenericAfterUpload(ctx, &fs)
	asserts.Equal(ErrPathNotExist, err)
	asserts.NoError(mock.ExpectationsWereMet())

	// 文件已存在
	mock.ExpectQuery("SELECT(.+)folders(.+)").WillReturnRows(
		mock.NewRows([]string{"name"}).AddRow("我的文件"),
	)
	mock.ExpectQuery("SELECT(.+)files(.+)").WillReturnRows(
		mock.NewRows([]string{"name"}).AddRow("test.txt"),
	)
	err = GenericAfterUpload(ctx, &fs)
	asserts.Equal(ErrFileExisted, err)
	asserts.NoError(mock.ExpectationsWereMet())

	// 插入失败
	mock.ExpectQuery("SELECT(.+)folders(.+)").WillReturnRows(
		mock.NewRows([]string{"name"}).AddRow("我的文件"),
	)
	mock.ExpectQuery("SELECT(.+)files(.+)").WillReturnError(errors.New("not found"))
	mock.ExpectBegin()
	mock.ExpectExec("INSERT(.+)files(.+)").WillReturnError(errors.New("error"))
	mock.ExpectRollback()

	err = GenericAfterUpload(ctx, &fs)
	asserts.Equal(ErrInsertFileRecord, err)
	asserts.NoError(mock.ExpectationsWereMet())

}
