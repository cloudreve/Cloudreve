package task

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestCompressTask_Props(t *testing.T) {
	asserts := assert.New(t)
	task := &CompressTask{
		User: &model.User{},
	}
	asserts.NotEmpty(task.Props())
	asserts.Equal(CompressTaskType, task.Type())
	asserts.EqualValues(0, task.Creator())
	asserts.Nil(task.Model())
}

func TestCompressTask_SetStatus(t *testing.T) {
	asserts := assert.New(t)
	task := &CompressTask{
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

func TestCompressTask_SetError(t *testing.T) {
	asserts := assert.New(t)
	task := &CompressTask{
		User: &model.User{},
		TaskModel: &model.Task{
			Model: gorm.Model{ID: 1},
		},
		zipPath: "test/TestCompressTask_SetError",
	}
	zipFile, _ := util.CreatNestedFile("test/TestCompressTask_SetError")
	zipFile.Close()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	task.SetErrorMsg("error")
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.False(util.Exists("test/TestCompressTask_SetError"))
	asserts.Equal("error", task.GetError().Msg)
}

func TestCompressTask_Do(t *testing.T) {
	asserts := assert.New(t)
	task := &CompressTask{
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

	// 压缩出错
	{
		task.User = &model.User{
			Policy: model.Policy{
				Type: "mock",
			},
		}
		task.TaskProps.Dirs = []uint{1}
		// 更新进度
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1,
			1))
		mock.ExpectCommit()
		// 查找目录
		mock.ExpectQuery("SELECT(.+)").WillReturnError(errors.New("error"))
		// 更新错误
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1,
			1))
		mock.ExpectCommit()
		task.Do()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NotEmpty(task.GetError().Msg)
	}

	// 上传出错
	{
		task.User = &model.User{
			Policy: model.Policy{
				Type:    "mock",
				MaxSize: 1,
			},
		}
		task.TaskProps.Dirs = []uint{1}
		cache.Set("setting_temp_path", "test", 0)
		// 更新进度
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1,
			1))
		mock.ExpectCommit()
		// 查找目录
		mock.ExpectQuery("SELECT(.+)folders").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		// 查找文件
		mock.ExpectQuery("SELECT(.+)files").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		// 查找子文件
		mock.ExpectQuery("SELECT(.+)files").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		// 查找子目录
		mock.ExpectQuery("SELECT(.+)folders").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		// 更新错误
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1,
			1))
		mock.ExpectCommit()
		task.Do()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NotEmpty(task.GetError().Msg)
		asserts.True(util.IsEmpty("test/compress"))
	}
}

func TestNewCompressTask(t *testing.T) {
	asserts := assert.New(t)

	// 成功
	{
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		job, err := NewCompressTask(&model.User{}, "/", []uint{12}, []uint{})
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NotNil(job)
		asserts.NoError(err)
	}

	// 失败
	{
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnError(errors.New("error"))
		mock.ExpectRollback()
		job, err := NewCompressTask(&model.User{}, "/", []uint{12}, []uint{})
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Nil(job)
		asserts.Error(err)
	}
}

func TestNewCompressTaskFromModel(t *testing.T) {
	asserts := assert.New(t)

	// 成功
	{
		mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		job, err := NewCompressTaskFromModel(&model.Task{Props: "{}"})
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.NotNil(job)
	}

	// JSON解析失败
	{
		mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		job, err := NewCompressTaskFromModel(&model.Task{Props: ""})
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.Nil(job)
	}
}
