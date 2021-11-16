package cluster

import (
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/balancer"
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
	a := assert.New(t)
	p := &NodePool{}
	n := &MasterNode{Model: &model.Node{}}
	p.Init()
	p.inactive[1] = n

	p.nodeStatusChange(true, 1)
	a.Len(p.inactive, 0)
	a.Equal(n, p.active[1])

	p.nodeStatusChange(false, 1)
	a.Len(p.active, 0)
	a.Equal(n, p.inactive[1])

	p.nodeStatusChange(false, 1)
	a.Len(p.active, 0)
	a.Equal(n, p.inactive[1])
}

func TestNodePool_Add(t *testing.T) {
	a := assert.New(t)
	p := &NodePool{}
	p.Init()

	// new node
	{
		p.Add(&model.Node{})
		a.Len(p.active, 1)
	}

	// old node
	{
		p.inactive[0] = p.active[0]
		delete(p.active, 0)
		p.Add(&model.Node{})
		a.Len(p.active, 0)
		a.Len(p.inactive, 1)
	}
}

func TestNodePool_Delete(t *testing.T) {
	a := assert.New(t)
	p := &NodePool{}
	p.Init()

	// active
	{
		mockNode := &nodeMock{}
		mockNode.On("Kill")
		p.active[0] = mockNode
		p.Delete(0)
		a.Len(p.active, 0)
		a.Len(p.inactive, 0)
		mockNode.AssertExpectations(t)
	}

	p.Init()

	// inactive
	{
		mockNode := &nodeMock{}
		mockNode.On("Kill")
		p.inactive[0] = mockNode
		p.Delete(0)
		a.Len(p.active, 0)
		a.Len(p.inactive, 0)
		mockNode.AssertExpectations(t)
	}
}

func TestNodePool_BalanceNodeByFeature(t *testing.T) {
	a := assert.New(t)
	p := &NodePool{}
	p.Init()

	// success
	{
		p.featureMap["test"] = []Node{&MasterNode{}}
		err, res := p.BalanceNodeByFeature("test", balancer.NewBalancer("round-robin"))
		a.NoError(err)
		a.Equal(p.featureMap["test"][0], res)
	}

	// NoNodes
	{
		p.featureMap["test"] = []Node{}
		err, res := p.BalanceNodeByFeature("test", balancer.NewBalancer("round-robin"))
		a.Error(err)
		a.Nil(res)
	}

	// No match feature
	{
		err, res := p.BalanceNodeByFeature("test2", balancer.NewBalancer("round-robin"))
		a.Error(err)
		a.Nil(res)
	}
}
