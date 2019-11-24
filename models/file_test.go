package model

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetFileByPathAndName(t *testing.T) {
	asserts := assert.New(t)

	fileRows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "1.cia")
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(fileRows)
	file, _ := GetFileByPathAndName("/", "1.cia", 1)
	asserts.Equal("1.cia", file.Name)
	asserts.NoError(mock.ExpectationsWereMet())
}

func TestFile_Create(t *testing.T) {
	asserts := assert.New(t)
	file := File{
		Name: "123",
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT(.+)").WillReturnResult(sqlmock.NewResult(5, 1))
	mock.ExpectCommit()
	fileID, err := file.Create()
	asserts.NoError(err)
	asserts.Equal(uint(5), fileID)
	asserts.Equal(uint(5), file.ID)
	asserts.NoError(mock.ExpectationsWereMet())

	mock.ExpectBegin()
	mock.ExpectExec("INSERT(.+)").WillReturnError(errors.New("error"))
	mock.ExpectRollback()
	fileID, err = file.Create()
	asserts.Error(err)
	asserts.NoError(mock.ExpectationsWereMet())

}

func TestFolder_GetChildFile(t *testing.T) {
	asserts := assert.New(t)
	folder := &Folder{
		Model: gorm.Model{
			ID: 1,
		},
	}

	// 找不到
	mock.ExpectQuery("SELECT(.+)folder_id(.+)").WithArgs(1).WillReturnError(errors.New("error"))
	files, err := folder.GetChildFile()
	asserts.Error(err)
	asserts.Len(files, 0)
	asserts.NoError(mock.ExpectationsWereMet())

	// 找到了
	mock.ExpectQuery("SELECT(.+)folder_id(.+)").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"name", "id"}).AddRow("1.txt", 1).AddRow("2.txt", 2))
	files, err = folder.GetChildFile()
	asserts.NoError(err)
	asserts.Len(files, 2)
	asserts.NoError(mock.ExpectationsWereMet())

}
