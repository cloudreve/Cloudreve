package model

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
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
