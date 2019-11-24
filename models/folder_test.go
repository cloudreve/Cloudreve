package model

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetFolderByPath(t *testing.T) {
	asserts := assert.New(t)

	folderRows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "test")
	mock.ExpectQuery("^SELECT (.+)").WillReturnRows(folderRows)
	folder, _ := GetFolderByPath("/测试/test", 1)
	asserts.Equal("test", folder.Name)
	asserts.NoError(mock.ExpectationsWereMet())

	folderRows = sqlmock.NewRows([]string{"id", "name"})
	mock.ExpectQuery("^SELECT (.+)").WillReturnRows(folderRows)
	folder, err := GetFolderByPath("/测试/test", 1)
	asserts.Error(err)
	asserts.NoError(mock.ExpectationsWereMet())
}

func TestFolder_Create(t *testing.T) {
	asserts := assert.New(t)
	folder := &Folder{
		Name: "new folder",
	}

	// 插入成功
	mock.ExpectBegin()
	mock.ExpectExec("INSERT(.+)").WillReturnResult(sqlmock.NewResult(5, 1))
	mock.ExpectCommit()
	fid, err := folder.Create()
	asserts.NoError(err)
	asserts.Equal(uint(5), fid)
	asserts.NoError(mock.ExpectationsWereMet())

	// 插入失败
	mock.ExpectBegin()
	mock.ExpectExec("INSERT(.+)").WillReturnError(errors.New("error"))
	mock.ExpectRollback()
	fid, err = folder.Create()
	asserts.Error(err)
	asserts.Equal(uint(0), fid)
	asserts.NoError(mock.ExpectationsWereMet())
}

func TestFolder_GetChildFolder(t *testing.T) {
	asserts := assert.New(t)
	folder := &Folder{
		Model: gorm.Model{
			ID: 1,
		},
	}

	// 找不到
	mock.ExpectQuery("SELECT(.+)parent_id(.+)").WithArgs(1).WillReturnError(errors.New("error"))
	files, err := folder.GetChildFolder()
	asserts.Error(err)
	asserts.Len(files, 0)
	asserts.NoError(mock.ExpectationsWereMet())

	// 找到了
	mock.ExpectQuery("SELECT(.+)parent_id(.+)").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"name", "id"}).AddRow("1.txt", 1).AddRow("2.txt", 2))
	files, err = folder.GetChildFolder()
	asserts.NoError(err)
	asserts.Len(files, 2)
	asserts.NoError(mock.ExpectationsWereMet())
}
