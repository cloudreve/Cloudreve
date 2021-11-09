package model

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTask_Create(t *testing.T) {
	asserts := assert.New(t)
	// 成功
	{
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		task := Task{Props: "1"}
		id, err := task.Create()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.EqualValues(1, id)
	}

	// 失败
	{
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnError(errors.New("error"))
		mock.ExpectRollback()
		task := Task{Props: "1"}
		id, err := task.Create()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.EqualValues(0, id)
	}
}

func TestTask_SetError(t *testing.T) {
	asserts := assert.New(t)
	task := Task{
		Model: gorm.Model{ID: 1},
	}
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	asserts.NoError(task.SetError("error"))
	asserts.NoError(mock.ExpectationsWereMet())
}

func TestTask_SetStatus(t *testing.T) {
	asserts := assert.New(t)
	task := Task{
		Model: gorm.Model{ID: 1},
	}
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	asserts.NoError(task.SetStatus(1))
	asserts.NoError(mock.ExpectationsWereMet())
}

func TestTask_SetProgress(t *testing.T) {
	asserts := assert.New(t)
	task := Task{
		Model: gorm.Model{ID: 1},
	}
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	asserts.NoError(task.SetProgress(1))
	asserts.NoError(mock.ExpectationsWereMet())
}

func TestGetTasksByID(t *testing.T) {
	asserts := assert.New(t)
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	res, err := GetTasksByID(1)
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.NoError(err)
	asserts.EqualValues(1, res.ID)
}

func TestListTasks(t *testing.T) {
	asserts := assert.New(t)

	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))

	res, total := ListTasks(1, 1, 10, "")
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.EqualValues(5, total)
	asserts.Len(res, 1)
}

func TestGetTasksByStatus(t *testing.T) {
	a := assert.New(t)

	mock.ExpectQuery("SELECT(.+)").
		WithArgs(1, 2).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	res := GetTasksByStatus(1, 2)
	a.NoError(mock.ExpectationsWereMet())
	a.Len(res, 1)
}
