package task

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

var mock sqlmock.Sqlmock

// TestMain 初始化数据库Mock
func TestMain(m *testing.M) {
	var db *sql.DB
	var err error
	db, mock, err = sqlmock.New()
	if err != nil {
		panic("An error was not expected when opening a stub database connection")
	}
	model.DB, _ = gorm.Open("mysql", db)
	defer db.Close()
	m.Run()
}

func TestInit(t *testing.T) {
	asserts := assert.New(t)
	cache.Set("setting_max_worker_num", "10", 0)
	mock.ExpectQuery("SELECT(.+)").WithArgs(Queued).WillReturnRows(sqlmock.NewRows([]string{"type"}).AddRow(-1))
	Init()
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.Len(TaskPoll.idleWorker, 10)
}

func TestPool_Submit(t *testing.T) {
	asserts := assert.New(t)
	pool := &Pool{
		idleWorker: make(chan int, 1),
	}
	pool.Add(1)
	job := &MockJob{
		DoFunc: func() {

		},
	}
	asserts.NotPanics(func() {
		pool.Submit(job)
	})
}
