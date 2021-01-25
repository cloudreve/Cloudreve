package scripts

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUserStorageCalibration_Run(t *testing.T) {
	asserts := assert.New(t)
	script := UserStorageCalibration(0)

	// 容量异常
	{
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "email", "storage"}).AddRow(1, "a@a.com", 10))
		mock.ExpectQuery("SELECT(.+)files(.+)").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(11))
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		script.Run(context.Background())
		asserts.NoError(mock.ExpectationsWereMet())
	}

	// 容量正常
	{
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "email", "storage"}).AddRow(1, "a@a.com", 10))
		mock.ExpectQuery("SELECT(.+)files(.+)").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(10))
		script.Run(context.Background())
		asserts.NoError(mock.ExpectationsWereMet())
	}
}
