package scripts

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"testing"
)

var mock sqlmock.Sqlmock
var mockDB *gorm.DB

// TestMain 初始化数据库Mock
func TestMain(m *testing.M) {
	var db *sql.DB
	var err error
	db, mock, err = sqlmock.New()
	if err != nil {
		panic("An error was not expected when opening a stub database connection")
	}
	model.DB, _ = gorm.Open("mysql", db)
	mockDB = model.DB
	defer db.Close()
	m.Run()
}

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
