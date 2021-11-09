package model

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTag_Create(t *testing.T) {
	asserts := assert.New(t)
	tag := Tag{}

	// 成功
	{
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		id, err := tag.Create()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.EqualValues(1, id)
	}

	// 失败
	{
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnError(errors.New("error"))
		mock.ExpectRollback()
		id, err := tag.Create()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.EqualValues(0, id)
	}
}

func TestDeleteTagByID(t *testing.T) {
	asserts := assert.New(t)
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := DeleteTagByID(1, 2)
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.NoError(err)
}

func TestGetTagsByUID(t *testing.T) {
	asserts := assert.New(t)
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	res, err := GetTagsByUID(1)
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.NoError(err)
	asserts.Len(res, 1)
}

func TestGetTagsByID(t *testing.T) {
	asserts := assert.New(t)
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("tag"))
	res, err := GetTagsByUID(1)
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.NoError(err)
	asserts.EqualValues("tag", res[0].Name)
}
