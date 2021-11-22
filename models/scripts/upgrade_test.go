package scripts

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpgradeTo340_Run(t *testing.T) {
	a := assert.New(t)
	script := UpgradeTo340(0)

	// skip
	{
		mock.ExpectQuery("SELECT(.+)settings").WillReturnRows(sqlmock.NewRows([]string{"name"}))
		script.Run(context.Background())
		a.NoError(mock.ExpectationsWereMet())
	}

	// node not found
	{
		mock.ExpectQuery("SELECT(.+)settings").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("1"))
		mock.ExpectQuery("SELECT(.+)nodes").WillReturnRows(sqlmock.NewRows([]string{"id"}))
		script.Run(context.Background())
		a.NoError(mock.ExpectationsWereMet())
	}

	// success
	{
		mock.ExpectQuery("SELECT(.+)settings").WillReturnRows(sqlmock.NewRows([]string{"name", "value"}).
			AddRow("aria2_rpcurl", "expected_aria2_rpcurl").
			AddRow("aria2_interval", "expected_aria2_interval").
			AddRow("aria2_temp_path", "expected_aria2_temp_path").
			AddRow("aria2_token", "expected_aria2_token").
			AddRow("aria2_options", "{}"))

		mock.ExpectQuery("SELECT(.+)nodes").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		script.Run(context.Background())
		a.NoError(mock.ExpectationsWereMet())
	}

	// failed
	{
		mock.ExpectQuery("SELECT(.+)settings").WillReturnRows(sqlmock.NewRows([]string{"name", "value"}).
			AddRow("aria2_rpcurl", "expected_aria2_rpcurl").
			AddRow("aria2_interval", "expected_aria2_interval").
			AddRow("aria2_temp_path", "expected_aria2_temp_path").
			AddRow("aria2_token", "expected_aria2_token").
			AddRow("aria2_options", "{}"))

		mock.ExpectQuery("SELECT(.+)nodes").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnError(errors.New("error"))
		mock.ExpectRollback()
		script.Run(context.Background())
		a.NoError(mock.ExpectationsWereMet())
	}
}
