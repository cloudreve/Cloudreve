package model

import (
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
