package model

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetNodeByID(t *testing.T) {
	a := assert.New(t)
	mock.ExpectQuery("SELECT(.+)nodes").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	res, err := GetNodeByID(1)
	a.NoError(err)
	a.EqualValues(1, res.ID)
	a.NoError(mock.ExpectationsWereMet())
}

func TestGetNodesByStatus(t *testing.T) {
	a := assert.New(t)
	mock.ExpectQuery("SELECT(.+)nodes").WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(NodeActive))
	res, err := GetNodesByStatus(NodeActive)
	a.NoError(err)
	a.Len(res, 1)
	a.EqualValues(NodeActive, res[0].Status)
	a.NoError(mock.ExpectationsWereMet())
}

func TestNode_AfterFind(t *testing.T) {
	a := assert.New(t)
	node := &Node{}

	// No aria2 options
	{
		a.NoError(node.AfterFind())
	}

	// with aria2 options
	{
		node.Aria2Options = `{"timeout":1}`
		a.NoError(node.AfterFind())
		a.Equal(1, node.Aria2OptionsSerialized.Timeout)
	}
}

func TestNode_BeforeSave(t *testing.T) {
	a := assert.New(t)
	node := &Node{}

	node.Aria2OptionsSerialized.Timeout = 1
	a.NoError(node.BeforeSave())
	a.Contains(node.Aria2Options, "1")
}

func TestNode_SetStatus(t *testing.T) {
	a := assert.New(t)
	node := &Node{}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE(.+)nodes").WithArgs(NodeActive, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	a.NoError(node.SetStatus(NodeActive))
	a.Equal(NodeActive, node.Status)
	a.NoError(mock.ExpectationsWereMet())
}
