package filesystem

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestFileSystem_IsFileExist(t *testing.T) {
	asserts := assert.New(t)
	fs := &FileSystem{User: &model.User{
		Model: gorm.Model{
			ID: 1,
		},
	}}

	// 存在
	{
		path := "/1.txt"
		// 根目录
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery("SELECT(.+)").WithArgs(1, "1.txt").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "1.txt"))
		exist, file := fs.IsFileExist(path)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.True(exist)
		asserts.Equal(uint(1), file.ID)
	}

	// 文件不存在
	{
		path := "/1.txt"
		// 根目录
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery("SELECT(.+)").WithArgs(1, "1.txt").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
		exist, _ := fs.IsFileExist(path)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.False(exist)
	}

	// 父目录不存在
	{
		path := "/1.txt"
		// 根目录
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		exist, _ := fs.IsFileExist(path)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.False(exist)
	}
}

func TestFileSystem_IsPathExist(t *testing.T) {
	asserts := assert.New(t)
	fs := &FileSystem{User: &model.User{
		Model: gorm.Model{
			ID: 1,
		},
	}}

	// 查询根目录
	{
		path := "/"
		// 根目录
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		exist, folder := fs.IsPathExist(path)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.True(exist)
		asserts.Equal(uint(1), folder.ID)
	}

	// 深层路径
	{
		path := "/1/2/3"
		// 根目录
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(1, 1))
		// 1
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1, 1, "1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(2, 1))
		// 2
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(2, 1, "2").
			WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(3, 1))
		// 3
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(3, 1, "3").
			WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(4, 1))
		exist, folder := fs.IsPathExist(path)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.True(exist)
		asserts.Equal(uint(4), folder.ID)
	}

	// 深层路径 重设根目录为/1
	{
		path := "/2/3"
		fs.Root = &model.Folder{Name: "1", Model: gorm.Model{ID: 2}, OwnerID: 1}
		// 2
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(2, 1, "2").
			WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(3, 1))
		// 3
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(3, 1, "3").
			WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(4, 1))
		exist, folder := fs.IsPathExist(path)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.True(exist)
		asserts.Equal(uint(4), folder.ID)
		fs.Root = nil
	}

	// 深层 不存在
	{
		path := "/1/2/3"
		// 根目录
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(1, 1))
		// 1
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1, 1, "1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(2, 1))
		// 2
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(2, 1, "2").
			WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(3, 1))
		// 3
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(3, 1, "3").
			WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}))
		exist, folder := fs.IsPathExist(path)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.False(exist)
		asserts.Nil(folder)
	}

}

func TestFileSystem_IsChildFileExist(t *testing.T) {
	asserts := assert.New(t)
	fs := &FileSystem{User: &model.User{
		Model: gorm.Model{
			ID: 1,
		},
	}}
	folder := model.Folder{
		Model:    gorm.Model{ID: 1},
		Name:     "123",
		Position: "/",
	}

	mock.ExpectQuery("SELECT(.+)").
		WithArgs(1, "321").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(2, "321"))
	exist, childFile := fs.IsChildFileExist(&folder, "321")
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.True(exist)
	asserts.Equal("/123", childFile.Position)
}
