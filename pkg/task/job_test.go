package task

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/stretchr/testify/assert"
)

func TestRecord(t *testing.T) {
	asserts := assert.New(t)
	job := &TransferTask{
		User: &model.User{Policy: model.Policy{Type: "unknown"}},
	}
	mock.ExpectBegin()
	mock.ExpectExec("INSERT(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	_, err := Record(job)
	asserts.NoError(err)
}

func TestResume(t *testing.T) {
	asserts := assert.New(t)

	// 没有任务
	{
		mock.ExpectQuery("SELECT(.+)").WithArgs(Queued).WillReturnRows(sqlmock.NewRows([]string{"type"}))
		Resume()
		asserts.NoError(mock.ExpectationsWereMet())
	}
}

func TestGetJobFromModel(t *testing.T) {
	asserts := assert.New(t)

	// CompressTaskType
	{
		task := &model.Task{
			Status: 0,
			Type:   CompressTaskType,
		}
		mock.ExpectQuery("SELECT(.+)users(.+)").WillReturnError(errors.New("error"))
		job, err := GetJobFromModel(task)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Nil(job)
		asserts.Error(err)
	}
	// DecompressTaskType
	{
		task := &model.Task{
			Status: 0,
			Type:   DecompressTaskType,
		}
		mock.ExpectQuery("SELECT(.+)users(.+)").WillReturnError(errors.New("error"))
		job, err := GetJobFromModel(task)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Nil(job)
		asserts.Error(err)
	}
	// TransferTaskType
	{
		task := &model.Task{
			Status: 0,
			Type:   TransferTaskType,
		}
		mock.ExpectQuery("SELECT(.+)users(.+)").WillReturnError(errors.New("error"))
		job, err := GetJobFromModel(task)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Nil(job)
		asserts.Error(err)
	}
}
