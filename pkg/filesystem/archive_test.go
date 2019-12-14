package filesystem

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/cache"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"testing"
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

		zipFile, err := fs.Compress(ctx, []uint{1}, []uint{1})
		asserts.NoError(err)
		asserts.NotEmpty(zipFile)
		asserts.Contains(zipFile, "archive_")
		asserts.Contains(zipFile, "tests")
	}
}
