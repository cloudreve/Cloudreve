package model

import (
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
