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

type TestScript int

func (script TestScript) Run(ctx context.Context) {

}

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

func TestRunDBScript(t *testing.T) {
	asserts := assert.New(t)
	register("test", TestScript(0))

	// 不存在
	{
		asserts.Error(RunDBScript("else", context.Background()))
	}

	// 存在
	{
		asserts.NoError(RunDBScript("test", context.Background()))
	}
}
