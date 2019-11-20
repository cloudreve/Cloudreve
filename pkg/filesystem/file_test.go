package filesystem

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem/local"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFileSystem_AddFile(t *testing.T) {
	asserts := assert.New(t)
	file := local.FileData{
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
