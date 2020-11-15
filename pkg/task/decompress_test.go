package task

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestDecompressTask_Props(t *testing.T) {
	asserts := assert.New(t)
	task := &DecompressTask{
		User: &model.User{},
	}
	asserts.NotEmpty(task.Props())
	asserts.Equal(DecompressTaskType, task.Type())
	asserts.EqualValues(0, task.Creator())
	asserts.Nil(task.Model())
}

func TestDecompressTask_SetStatus(t *testing.T) {
	asserts := assert.New(t)
	task := &DecompressTask{
		User: &model.User{},
		TaskModel: &model.Task{
			Model: gorm.Model{ID: 1},
		},
	}
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	task.SetStatus(3)
	asserts.NoError(mock.ExpectationsWereMet())
}

func TestDecompressTask_SetError(t *testing.T) {
	asserts := assert.New(t)
	task := &DecompressTask{
		User: &model.User{},
		TaskModel: &model.Task{
			Model: gorm.Model{ID: 1},
		},
		zipPath: "test/TestCompressTask_SetError",
	}
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	task.SetErrorMsg("error", nil)
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.Equal("error", task.GetError().Msg)
}

func TestDecompressTask_Do(t *testing.T) {
	asserts := assert.New(t)
	task := &DecompressTask{
		TaskModel: &model.Task{
			Model: gorm.Model{ID: 1},
		},
	}

	// 无法创建文件系统
	{
		task.User = &model.User{
			Policy: model.Policy{
				Type: "unknown",
			},
		}
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1,
			1))
		mock.ExpectCommit()
		task.Do()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NotEmpty(task.GetError().Msg)
	}

	// 压缩文件不存在
	{
		task.User = &model.User{
			Policy: model.Policy{
				Type: "mock",
			},
		}
		task.TaskProps.Src = "test"
		task.Do()
		asserts.NotEmpty(task.GetError().Msg)
	}
}

func TestNewDecompressTask(t *testing.T) {
	asserts := assert.New(t)

	// 成功
	{
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		job, err := NewDecompressTask(&model.User{}, "/", "/")
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NotNil(job)
		asserts.NoError(err)
	}

	// 失败
	{
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnError(errors.New("error"))
		mock.ExpectRollback()
		job, err := NewDecompressTask(&model.User{}, "/", "/")
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Nil(job)
		asserts.Error(err)
	}
}

func TestNewDecompressTaskFromModel(t *testing.T) {
	asserts := assert.New(t)

	// 成功
	{
		mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		job, err := NewDecompressTaskFromModel(&model.Task{Props: "{}"})
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.NotNil(job)
	}

	// JSON解析失败
	{
		mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		job, err := NewDecompressTaskFromModel(&model.Task{Props: ""})
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.Nil(job)
	}
}
