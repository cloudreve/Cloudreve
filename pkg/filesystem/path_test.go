package filesystem

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestFileSystem_IsFileExist(t *testing.T) {
	asserts := assert.New(t)
	fs := &FileSystem{User: &model.User{
		Model: gorm.Model{
			ID: 1,
		},
	}}

	// 存在
	mock.ExpectQuery("SELECT(.+)").WithArgs(uint(1), "/s", "1.txt").WillReturnRows(
		sqlmock.NewRows([]string{"Name"}).AddRow("s"),
	)
	testResult, _ := fs.IsFileExist("/s/1.txt")
	asserts.True(testResult)
	asserts.NoError(mock.ExpectationsWereMet())

	// 不存在
	mock.ExpectQuery("SELECT(.+)").WithArgs(uint(1), "/ss/dfsd", "1.txt").WillReturnRows(
		sqlmock.NewRows([]string{"Name"}),
	)
	testResult, _ = fs.IsFileExist("/ss/dfsd/1.txt")
	asserts.False(testResult)
	asserts.NoError(mock.ExpectationsWereMet())
}

func TestFileSystem_IsPathExist(t *testing.T) {
	asserts := assert.New(t)
	fs := &FileSystem{User: &model.User{
		Model: gorm.Model{
			ID: 1,
		},
	}}

	// 存在
	mock.ExpectQuery("SELECT(.+)").WithArgs(uint(1), "/s/sda").WillReturnRows(
		sqlmock.NewRows([]string{"name"}).AddRow("sda"),
	)
	testResult, folder := fs.IsPathExist("/s/sda")
	asserts.True(testResult)
	asserts.Equal("sda", folder.Name)
	asserts.NoError(mock.ExpectationsWereMet())

	// 不存在
	mock.ExpectQuery("SELECT(.+)").WithArgs(uint(1), "/s/sda").WillReturnRows(
		sqlmock.NewRows([]string{"name"}),
	)
	testResult, _ = fs.IsPathExist("/s/sda")
	asserts.False(testResult)
	asserts.NoError(mock.ExpectationsWereMet())
}

func TestFileSystem_List(t *testing.T) {
	asserts := assert.New(t)
	fs := &FileSystem{User: &model.User{
		Model: gorm.Model{
			ID: 1,
		},
	}}
	ctx := context.Background()

	// 成功，子目录包含文件和路径，不使用路径处理钩子
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(5, "folder"))
	mock.ExpectQuery("SELECT(.+)folder(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(6, "sub_folder1").AddRow(7, "sub_folder2"))
	mock.ExpectQuery("SELECT(.+)file(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(6, "sub_file1.txt").AddRow(7, "sub_file2.txt"))
	objects, err := fs.List(ctx, "/folder", nil)
	asserts.Len(objects, 4)
	asserts.NoError(err)
	asserts.NoError(mock.ExpectationsWereMet())

	// 成功，子目录包含文件和路径，使用路径处理钩子
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(5, "folder"))
	mock.ExpectQuery("SELECT(.+)folder(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "position"}).AddRow(6, "sub_folder1", "/folder").AddRow(7, "sub_folder2", "/folder"))
	mock.ExpectQuery("SELECT(.+)file(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "dir"}).AddRow(6, "sub_file1.txt", "/folder").AddRow(7, "sub_file2.txt", "/folder"))
	objects, err = fs.List(ctx, "/folder", func(s string) string {
		return "prefix" + s
	})
	asserts.Len(objects, 4)
	asserts.NoError(err)
	asserts.NoError(mock.ExpectationsWereMet())
	for _, value := range objects {
		asserts.Contains(value.Path, "prefix/")
	}

	// 成功，子目录包含路径，使用路径处理钩子
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(5, "folder"))
	mock.ExpectQuery("SELECT(.+)folder(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "position"}))
	mock.ExpectQuery("SELECT(.+)file(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "dir"}).AddRow(6, "sub_file1.txt", "/folder").AddRow(7, "sub_file2.txt", "/folder"))
	objects, err = fs.List(ctx, "/folder", func(s string) string {
		return "prefix" + s
	})
	asserts.Len(objects, 2)
	asserts.NoError(err)
	asserts.NoError(mock.ExpectationsWereMet())
	for _, value := range objects {
		asserts.Contains(value.Path, "prefix/")
	}

	// 成功，子目录下为空，使用路径处理钩子
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(5, "folder"))
	mock.ExpectQuery("SELECT(.+)folder(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "position"}))
	mock.ExpectQuery("SELECT(.+)file(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "dir"}))
	objects, err = fs.List(ctx, "/folder", func(s string) string {
		return "prefix" + s
	})
	asserts.Len(objects, 0)
	asserts.NoError(err)
	asserts.NoError(mock.ExpectationsWereMet())

	// 成功，子目录路径不存在
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	objects, err = fs.List(ctx, "/folder", func(s string) string {
		return "prefix" + s
	})
	asserts.Len(objects, 0)
	asserts.NoError(mock.ExpectationsWereMet())
}

func TestFileSystem_CreateDirectory(t *testing.T) {
	asserts := assert.New(t)
	fs := &FileSystem{User: &model.User{
		Model: gorm.Model{
			ID: 1,
		},
	}}
	ctx := context.Background()

	// 目录名非法
	err := fs.CreateDirectory(ctx, "/ad/a+?")
	asserts.Equal(ErrIllegalObjectName, err)

	// 父目录不存在
	mock.ExpectQuery("SELECT(.+)folders").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
	err = fs.CreateDirectory(ctx, "/ad/ab")
	asserts.Equal(ErrPathNotExist, err)
	asserts.NoError(mock.ExpectationsWereMet())

	// 存在同名文件
	mock.ExpectQuery("SELECT(.+)folders").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "ab"))
	mock.ExpectQuery("SELECT(.+)files").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "ab"))
	err = fs.CreateDirectory(ctx, "/ad/ab")
	asserts.Equal(ErrFileExisted, err)
	asserts.NoError(mock.ExpectationsWereMet())

	// 存在同名目录
	mock.ExpectQuery("SELECT(.+)folders").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "ab"))
	mock.ExpectQuery("SELECT(.+)folders").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "ab"))
	err = fs.CreateDirectory(ctx, "/ad/ab")
	asserts.Equal(ErrFolderExisted, err)
	asserts.NoError(mock.ExpectationsWereMet())

	// 成功创建
	mock.ExpectQuery("SELECT(.+)folders").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "ab"))
	mock.ExpectQuery("SELECT(.+)files").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
	mock.ExpectBegin()
	mock.ExpectExec("INSERT(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err = fs.CreateDirectory(ctx, "/ad/ab")
	asserts.NoError(err)
	asserts.NoError(mock.ExpectationsWereMet())
}

func TestFileSystem_ListDeleteFiles(t *testing.T) {
	conf.DatabaseConfig.Type = "mysql"
	asserts := assert.New(t)
	fs := &FileSystem{User: &model.User{
		Model: gorm.Model{
			ID: 1,
		},
	}}

	// 成功
	{
		mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "1.txt").AddRow(2, "2.txt"))
		err := fs.ListDeleteFiles(context.Background(), []string{"/"})
		asserts.NoError(err)
		asserts.NoError(mock.ExpectationsWereMet())
	}

	// 失败
	{
		mock.ExpectQuery("SELECT(.+)").WillReturnError(errors.New("error"))
		err := fs.ListDeleteFiles(context.Background(), []string{"/"})
		asserts.Error(err)
		asserts.Equal(serializer.CodeDBError, err.(serializer.AppError).Code)
		asserts.NoError(mock.ExpectationsWereMet())
	}
}

func TestFileSystem_ListDeleteDirs(t *testing.T) {
	conf.DatabaseConfig.Type = "mysql"
	asserts := assert.New(t)
	fs := &FileSystem{User: &model.User{
		Model: gorm.Model{
			ID: 1,
		},
	}}

	// 成功
	{
		mock.ExpectQuery("SELECT(.+)").
			WillReturnRows(
				sqlmock.NewRows([]string{"id"}).
					AddRow(1).
					AddRow(2).
					AddRow(3),
			)
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1, 2, 3).
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "name"}).
					AddRow(4, "1.txt").
					AddRow(5, "2.txt").
					AddRow(6, "3.txt"),
			)
		err := fs.ListDeleteDirs(context.Background(), []string{"/"})
		asserts.NoError(err)
		asserts.Len(fs.FileTarget, 3)
		asserts.Len(fs.DirTarget, 3)
		asserts.NoError(mock.ExpectationsWereMet())
	}

	// 检索文件发生错误
	{
		mock.ExpectQuery("SELECT(.+)").
			WillReturnRows(
				sqlmock.NewRows([]string{"id"}).
					AddRow(1).
					AddRow(2).
					AddRow(3),
			)
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1, 2, 3).
			WillReturnError(errors.New("error"))
		err := fs.ListDeleteDirs(context.Background(), []string{"/"})
		asserts.Error(err)
		asserts.Len(fs.DirTarget, 6)
		asserts.NoError(mock.ExpectationsWereMet())
	}
	// 检索目录发生错误
	{
		mock.ExpectQuery("SELECT(.+)").
			WillReturnError(errors.New("error"))
		err := fs.ListDeleteDirs(context.Background(), []string{"/"})
		asserts.Error(err)
		asserts.NoError(mock.ExpectationsWereMet())
	}
}

func TestFileSystem_Delete(t *testing.T) {
	conf.DatabaseConfig.Type = "mysql"
	asserts := assert.New(t)
	fs := &FileSystem{User: &model.User{
		Model: gorm.Model{
			ID: 1,
		},
		Storage: 3,
		Group:   model.Group{MaxStorage: 3},
	}}
	ctx := context.Background()

	// 全部未成功
	{
		// 列出要删除的目录
		mock.ExpectQuery("SELECT(.+)").
			WillReturnRows(
				sqlmock.NewRows([]string{"id"}).
					AddRow(1).
					AddRow(2).
					AddRow(3),
			)
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1, 2, 3).
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "name", "source_name", "policy_id", "size"}).
					AddRow(4, "1.txt", "1.txt", 2, 1),
			)
		// 查询顶级的文件
		mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "source_name", "policy_id", "size"}).AddRow(1, "1.txt", "1.txt", 1, 2))
		mock.ExpectQuery("SELECT(.+)files(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "policy_id", "source_name"}))
		// 查询上传策略
		mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(1, "local"))
		mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(1, "local"))
		// 删除文件记录
		mock.ExpectBegin()
		mock.ExpectExec("DELETE(.+)").
			WillReturnResult(sqlmock.NewResult(0, 3))
		mock.ExpectCommit()
		// 归还容量
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").
			WillReturnResult(sqlmock.NewResult(0, 3))
		mock.ExpectCommit()
		// 删除目录
		mock.ExpectBegin()
		mock.ExpectExec("DELETE(.+)").
			WillReturnResult(sqlmock.NewResult(0, 3))
		mock.ExpectCommit()

		err := fs.Delete(ctx, []string{"/"}, []string{"2.txt"})
		asserts.Error(err)
		asserts.Equal(203, err.(serializer.AppError).Code)
		asserts.Equal(uint64(0), fs.User.Storage)
	}
	// 全部成功
	{
		file, err := os.Create("1.txt")
		file2, err := os.Create("2.txt")
		file.Close()
		file2.Close()
		asserts.NoError(err)
		mock.ExpectQuery("SELECT(.+)").
			WillReturnRows(
				sqlmock.NewRows([]string{"id"}).
					AddRow(1).
					AddRow(2).
					AddRow(3),
			)
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1, 2, 3).
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "name", "source_name", "policy_id", "size"}).
					AddRow(4, "1.txt", "1.txt", 2, 1),
			)
		mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "source_name", "policy_id", "size"}).AddRow(1, "2.txt", "2.txt", 1, 2))
		mock.ExpectQuery("SELECT(.+)files(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "policy_id", "source_name"}))
		// 查询上传策略
		mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(1, "local"))
		mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(1, "local"))
		// 删除文件记录
		mock.ExpectBegin()
		mock.ExpectExec("DELETE(.+)").
			WillReturnResult(sqlmock.NewResult(0, 3))
		mock.ExpectCommit()
		// 归还容量
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").
			WillReturnResult(sqlmock.NewResult(0, 3))
		mock.ExpectCommit()
		// 删除目录
		mock.ExpectBegin()
		mock.ExpectExec("DELETE(.+)").
			WillReturnResult(sqlmock.NewResult(0, 3))
		mock.ExpectCommit()

		fs.FileTarget = []model.File{}
		fs.DirTarget = []model.Folder{}
		err = fs.Delete(ctx, []string{"/"}, []string{"2.txt"})
		asserts.NoError(err)
		asserts.Equal(uint64(0), fs.User.Storage)
	}

}

func TestFileSystem_Copy(t *testing.T) {
	asserts := assert.New(t)
	fs := &FileSystem{User: &model.User{
		Model: gorm.Model{
			ID: 1,
		},
		Storage: 3,
		Group:   model.Group{MaxStorage: 3},
	}}
	ctx := context.Background()

	// 目录不存在
	{
		mock.ExpectQuery("SELECT(.+)").WithArgs(uint(1), "/dst").WillReturnRows(
			sqlmock.NewRows([]string{"name"}),
		)
		mock.ExpectQuery("SELECT(.+)").WithArgs(uint(1), "/src").WillReturnRows(
			sqlmock.NewRows([]string{"name"}),
		)
		err := fs.Copy(ctx, []string{}, []string{}, "/src", "/dst")
		asserts.Equal(ErrPathNotExist, err)
		asserts.NoError(mock.ExpectationsWereMet())
	}

	// 复制目录出错
	{
		mock.ExpectQuery("SELECT(.+)").WithArgs(uint(1), "/src").WillReturnRows(
			sqlmock.NewRows([]string{"name"}).AddRow("123"),
		)
		mock.ExpectQuery("SELECT(.+)").WithArgs(uint(1), "/dst").WillReturnRows(
			sqlmock.NewRows([]string{"name"}).AddRow("123"),
		)
		err := fs.Copy(ctx, []string{"test"}, []string{}, "/src", "/dst")
		asserts.Error(err)
	}

}

func TestFileSystem_Move(t *testing.T) {
	conf.DatabaseConfig.Type = "mysql"
	asserts := assert.New(t)
	fs := &FileSystem{User: &model.User{
		Model: gorm.Model{
			ID: 1,
		},
		Storage: 3,
		Group:   model.Group{MaxStorage: 3},
	}}
	ctx := context.Background()

	// 目录不存在
	{
		mock.ExpectQuery("SELECT(.+)").WillReturnRows(
			sqlmock.NewRows([]string{"name"}),
		)
		err := fs.Move(ctx, []string{}, []string{}, "/src", "/dst")
		asserts.Equal(ErrPathNotExist, err)
		asserts.NoError(mock.ExpectationsWereMet())
	}

	// 移动目录出错
	{
		mock.ExpectQuery("SELECT(.+)").WithArgs(uint(1), "/src").WillReturnRows(
			sqlmock.NewRows([]string{"name"}).AddRow("123"),
		)
		mock.ExpectQuery("SELECT(.+)").WithArgs(uint(1), "/dst").WillReturnRows(
			sqlmock.NewRows([]string{"name"}).AddRow("123"),
		)
		err := fs.Move(ctx, []string{"test"}, []string{}, "/src", "/dst")
		asserts.Error(err)
	}
}
