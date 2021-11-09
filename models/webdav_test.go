package model

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWebdav_Create(t *testing.T) {
	asserts := assert.New(t)
	// 成功
	{
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		task := Webdav{}
		id, err := task.Create()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.EqualValues(1, id)
	}

	// 失败
	{
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnError(errors.New("error"))
		mock.ExpectRollback()
		task := Webdav{}
		id, err := task.Create()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.EqualValues(0, id)
	}
}

func TestGetWebdavByPassword(t *testing.T) {
	asserts := assert.New(t)
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	_, err := GetWebdavByPassword("e", 1)
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.Error(err)
}

func TestListWebDAVAccounts(t *testing.T) {
	asserts := assert.New(t)
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	res := ListWebDAVAccounts(1)
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.Len(res, 0)
}

func TestDeleteWebDAVAccountByID(t *testing.T) {
	asserts := assert.New(t)
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	DeleteWebDAVAccountByID(1, 1)
	asserts.NoError(mock.ExpectationsWereMet())
}
