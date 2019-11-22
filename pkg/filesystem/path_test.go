package filesystem

import (
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
	testResult := fs.IsFileExist("/s/1.txt")
	asserts.True(testResult)
	asserts.NoError(mock.ExpectationsWereMet())

	// 不存在
	mock.ExpectQuery("SELECT(.+)").WithArgs(uint(1), "/ss/dfsd", "1.txt").WillReturnRows(
		sqlmock.NewRows([]string{"Name"}),
	)
	testResult = fs.IsFileExist("/ss/dfsd/1.txt")
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
