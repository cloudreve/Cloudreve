package scripts

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResetAdminPassword_Run(t *testing.T) {
	asserts := assert.New(t)
	script := ResetAdminPassword(0)

	// 初始用户不存在
	{
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "email", "storage"}))
		asserts.Panics(func() {
			script.Run(context.Background())
		})
		asserts.NoError(mock.ExpectationsWereMet())
	}

	// 密码更新失败
	{
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "email", "storage"}).AddRow(1, "a@a.com", 10))
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnError(errors.New("error"))
		mock.ExpectRollback()
		asserts.Panics(func() {
			script.Run(context.Background())
		})
		asserts.NoError(mock.ExpectationsWereMet())
	}

	// 成功
	{
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "email", "storage"}).AddRow(1, "a@a.com", 10))
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		asserts.NotPanics(func() {
			script.Run(context.Background())
		})
		asserts.NoError(mock.ExpectationsWereMet())
	}
}
