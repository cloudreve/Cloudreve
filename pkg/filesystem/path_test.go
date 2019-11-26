package filesystem

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/HFO4/cloudreve/models"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
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
