package model

import (
	"github.com/DATA-DOG/go-sqlmock"
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
