package aria2

import (
	"database/sql"
	"github.com/cloudreve/Cloudreve/v3/pkg/mocks"
	"github.com/cloudreve/Cloudreve/v3/pkg/mq"
	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/jinzhu/gorm"
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
	a := assert.New(t)
	mockPool := &mocks.NodePoolMock{}
	mockPool.On("GetNodeByID", testMock.Anything).Return(nil)
	mockQueue := mq.NewMQ()

	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	Init(false, mockPool, mockQueue)
	a.NoError(mock.ExpectationsWereMet())
	mockPool.AssertExpectations(t)
}

func TestTestRPCConnection(t *testing.T) {
	a := assert.New(t)

	// url not legal
	{
		res, err := TestRPCConnection(string([]byte{0x7f}), "", 10)
		a.Error(err)
		a.Empty(res.Version)
	}

	// rpc failed
	{
		res, err := TestRPCConnection("ws://0.0.0.0", "", 0)
		a.Error(err)
		a.Empty(res.Version)
	}
}

func TestGetLoadBalancer(t *testing.T) {
	a := assert.New(t)
	a.NotPanics(func() {
		GetLoadBalancer()
	})
}
