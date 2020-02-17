package model

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestUser_GetAvailablePackSize(t *testing.T) {
	asserts := assert.New(t)
	user := User{
		Model: gorm.Model{ID: 1},
	}

	// 未命中缓存
	{
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(sqlmock.AnyArg(), 1).
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "size", "expired_time"}).
					AddRow(1, 10, time.Now().Add(time.Second*time.Duration(20))).
					AddRow(3, 0, nil).
					AddRow(2, 20, time.Now().Add(time.Second*time.Duration(10))),
			)
		total := user.GetAvailablePackSize()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.EqualValues(30, total)
	}

	// 命中缓存
	{
		total := user.GetAvailablePackSize()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.EqualValues(30, total)
	}

}

func TestStoragePack_Create(t *testing.T) {
	asserts := assert.New(t)
	pack := &StoragePack{}

	// 成功
	{
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		id, err := pack.Create()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Equal(uint(1), id)
		asserts.NoError(err)
	}

	// 失败
	{
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnError(errors.New("error"))
		mock.ExpectRollback()
		id, err := pack.Create()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Equal(uint(0), id)
		asserts.Error(err)
	}
}

func TestGetExpiredStoragePack(t *testing.T) {
	asserts := assert.New(t)
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	res := GetExpiredStoragePack()
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.Len(res, 0)
}

func TestStoragePack_Delete(t *testing.T) {
	asserts := assert.New(t)
	pack := &StoragePack{}
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	asserts.NoError(pack.Delete())
	asserts.NoError(mock.ExpectationsWereMet())

}
