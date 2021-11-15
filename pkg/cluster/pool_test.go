package cluster

import (
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"testing"
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

func TestInitFailed(t *testing.T) {
	a := assert.New(t)
	mock.ExpectQuery("SELECT(.+)").WillReturnError(errors.New("error"))
	Init()
	a.NoError(mock.ExpectationsWereMet())
}

func TestInitSuccess(t *testing.T) {
	a := assert.New(t)
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "aria2_enabled", "type"}).AddRow(1, true, model.MasterNodeType))
	Init()
	a.NoError(mock.ExpectationsWereMet())
}

func TestNodePool_GetNodeByID(t *testing.T) {
	a := assert.New(t)
	p := &NodePool{}
	p.Init()
	mockNode := &nodeMock{}

	// inactive
	{
		p.inactive[1] = mockNode
		a.Equal(mockNode, p.GetNodeByID(1))
	}

	// active
	{
		delete(p.inactive, 1)
		p.active[1] = mockNode
		a.Equal(mockNode, p.GetNodeByID(1))
	}
}

func TestNodePool_NodeStatusChange(t *testing.T) {

}
