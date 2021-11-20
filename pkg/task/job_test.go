package task

import (
	"errors"
	testMock "github.com/stretchr/testify/mock"
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

type taskPoolMock struct {
	testMock.Mock
}

func (t taskPoolMock) Add(num int) {
	t.Called(num)
}

func (t taskPoolMock) Submit(job Job) {
	t.Called(job)
}

func TestResume(t *testing.T) {
	asserts := assert.New(t)
	mockPool := taskPoolMock{}

	// 没有任务
	{
		mock.ExpectQuery("SELECT(.+)").WithArgs(Queued, Processing).WillReturnRows(sqlmock.NewRows([]string{"type"}))
		Resume(mockPool)
		asserts.NoError(mock.ExpectationsWereMet())
	}

	// 有任务, 类型未知
	{
		mock.ExpectQuery("SELECT(.+)").WithArgs(Queued, Processing).WillReturnRows(sqlmock.NewRows([]string{"type"}).AddRow(233))
		Resume(mockPool)
		asserts.NoError(mock.ExpectationsWereMet())
	}

	// 有任务
	{
		mockPool.On("Submit", testMock.Anything)
		mock.ExpectQuery("SELECT(.+)").WithArgs(Queued, Processing).WillReturnRows(sqlmock.NewRows([]string{"type", "props"}).AddRow(CompressTaskType, "{}"))
		mock.ExpectQuery("SELECT(.+)users").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery("SELECT(.+)policies").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		Resume(mockPool)
		asserts.NoError(mock.ExpectationsWereMet())
		mockPool.AssertExpectations(t)
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
