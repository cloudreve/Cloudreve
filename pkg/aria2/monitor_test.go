package aria2

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/rpc"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/task"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
)

type InstanceMock struct {
	testMock.Mock
}

func (m InstanceMock) CreateTask(task *model.Download, options map[string]interface{}) error {
	args := m.Called(task, options)
	return args.Error(0)
}

func (m InstanceMock) Status(task *model.Download) (rpc.StatusInfo, error) {
	args := m.Called(task)
	return args.Get(0).(rpc.StatusInfo), args.Error(1)
}

func (m InstanceMock) Cancel(task *model.Download) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m InstanceMock) Select(task *model.Download, files []int) error {
	args := m.Called(task, files)
	return args.Error(0)
}

func TestNewMonitor(t *testing.T) {
	asserts := assert.New(t)
	NewMonitor(&model.Download{GID: "gid"})
	_, ok := EventNotifier.Subscribes.Load("gid")
	asserts.True(ok)
}

func TestMonitor_Loop(t *testing.T) {
	asserts := assert.New(t)
	notifier := make(chan StatusEvent)
	MAX_RETRY = 0
	monitor := &Monitor{
		Task:     &model.Download{GID: "gid"},
		Interval: time.Duration(1) * time.Second,
		notifier: notifier,
	}
	asserts.NotPanics(func() {
		monitor.Loop()
	})
}

func TestMonitor_Update(t *testing.T) {
	asserts := assert.New(t)
	monitor := &Monitor{
		Task: &model.Download{
			GID:    "gid",
			Parent: "TestMonitor_Update",
		},
		Interval: time.Duration(1) * time.Second,
	}

	// 无法获取状态
	{
		MAX_RETRY = 1
		testInstance := new(InstanceMock)
		testInstance.On("Status", testMock.Anything).Return(rpc.StatusInfo{}, errors.New("error"))
		file, _ := util.CreatNestedFile("TestMonitor_Update/1")
		file.Close()
		Instance = testInstance
		asserts.False(monitor.Update())
		asserts.True(monitor.Update())
		testInstance.AssertExpectations(t)
		asserts.False(util.Exists("TestMonitor_Update"))
	}

	// 磁力链下载重定向
	{
		testInstance := new(InstanceMock)
		testInstance.On("Status", testMock.Anything).Return(rpc.StatusInfo{
			FollowedBy: []string{"1"},
		}, nil)
		monitor.Task.ID = 1
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		Instance = testInstance
		asserts.False(monitor.Update())
		asserts.NoError(mock.ExpectationsWereMet())
		testInstance.AssertExpectations(t)
		asserts.EqualValues("1", monitor.Task.GID)
	}

	// 无法更新任务信息
	{
		testInstance := new(InstanceMock)
		testInstance.On("Status", testMock.Anything).Return(rpc.StatusInfo{}, nil)
		monitor.Task.ID = 1
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnError(errors.New("error"))
		mock.ExpectRollback()
		Instance = testInstance
		asserts.True(monitor.Update())
		asserts.NoError(mock.ExpectationsWereMet())
		testInstance.AssertExpectations(t)
	}

	// 返回未知状态
	{
		testInstance := new(InstanceMock)
		testInstance.On("Status", testMock.Anything).Return(rpc.StatusInfo{Status: "?"}, nil)
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		Instance = testInstance
		asserts.True(monitor.Update())
		asserts.NoError(mock.ExpectationsWereMet())
		testInstance.AssertExpectations(t)
	}

	// 返回被取消状态
	{
		testInstance := new(InstanceMock)
		testInstance.On("Status", testMock.Anything).Return(rpc.StatusInfo{Status: "removed"}, nil)
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		Instance = testInstance
		asserts.True(monitor.Update())
		asserts.NoError(mock.ExpectationsWereMet())
		testInstance.AssertExpectations(t)
	}

	// 返回活跃状态
	{
		testInstance := new(InstanceMock)
		testInstance.On("Status", testMock.Anything).Return(rpc.StatusInfo{Status: "active"}, nil)
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		Instance = testInstance
		asserts.False(monitor.Update())
		asserts.NoError(mock.ExpectationsWereMet())
		testInstance.AssertExpectations(t)
	}

	// 返回错误状态
	{
		testInstance := new(InstanceMock)
		testInstance.On("Status", testMock.Anything).Return(rpc.StatusInfo{Status: "error"}, nil)
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		Instance = testInstance
		asserts.True(monitor.Update())
		asserts.NoError(mock.ExpectationsWereMet())
		testInstance.AssertExpectations(t)
	}

	// 返回完成
	{
		testInstance := new(InstanceMock)
		testInstance.On("Status", testMock.Anything).Return(rpc.StatusInfo{Status: "complete"}, nil)
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		Instance = testInstance
		asserts.True(monitor.Update())
		asserts.NoError(mock.ExpectationsWereMet())
		testInstance.AssertExpectations(t)
	}
}

func TestMonitor_UpdateTaskInfo(t *testing.T) {
	asserts := assert.New(t)
	monitor := &Monitor{
		Task: &model.Download{
			Model:  gorm.Model{ID: 1},
			GID:    "gid",
			Parent: "TestMonitor_UpdateTaskInfo",
		},
	}

	// 失败
	{
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnError(errors.New("error"))
		mock.ExpectRollback()
		err := monitor.UpdateTaskInfo(rpc.StatusInfo{})
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
	}

	// 更新成功，无需校验
	{
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		err := monitor.UpdateTaskInfo(rpc.StatusInfo{})
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
	}

	// 更新成功，大小改变，需要校验，校验失败
	{
		testInstance := new(InstanceMock)
		testInstance.On("Cancel", testMock.Anything).Return(nil)
		Instance = testInstance
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		err := monitor.UpdateTaskInfo(rpc.StatusInfo{TotalLength: "1"})
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		testInstance.AssertExpectations(t)
	}
}

func TestMonitor_ValidateFile(t *testing.T) {
	asserts := assert.New(t)
	monitor := &Monitor{
		Task: &model.Download{
			Model:  gorm.Model{ID: 1},
			GID:    "gid",
			Parent: "TestMonitor_ValidateFile",
		},
	}

	// 无法创建文件系统
	{
		monitor.Task.User = &model.User{
			Policy: model.Policy{
				Type: "unknown",
			},
		}
		asserts.Error(monitor.ValidateFile())
	}

	// 文件大小超出容量配额
	{
		cache.Set("pack_size_0", uint64(0), 0)
		monitor.Task.TotalSize = 11
		monitor.Task.User = &model.User{
			Policy: model.Policy{
				Type: "mock",
			},
			Group: model.Group{
				MaxStorage: 10,
			},
		}
		asserts.Equal(filesystem.ErrInsufficientCapacity, monitor.ValidateFile())
	}

	// 单文件大小超出容量配额
	{
		cache.Set("pack_size_0", uint64(0), 0)
		monitor.Task.TotalSize = 10
		monitor.Task.StatusInfo.Files = []rpc.FileInfo{
			{
				Selected: "true",
				Length:   "6",
			},
		}
		monitor.Task.User = &model.User{
			Policy: model.Policy{
				Type:    "mock",
				MaxSize: 5,
			},
			Group: model.Group{
				MaxStorage: 10,
			},
		}
		asserts.Equal(filesystem.ErrFileSizeTooBig, monitor.ValidateFile())
	}
}

func TestMonitor_Complete(t *testing.T) {
	asserts := assert.New(t)
	monitor := &Monitor{
		Task: &model.Download{
			Model:  gorm.Model{ID: 1},
			GID:    "gid",
			Parent: "TestMonitor_Complete",
			StatusInfo: rpc.StatusInfo{
				Files: []rpc.FileInfo{
					{
						Selected: "true",
						Path:     "TestMonitor_Complete",
					},
				},
			},
		},
	}

	cache.Set("setting_max_worker_num", "1", 0)
	mock.ExpectQuery("SELECT(.+)tasks").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	task.Init()
	mock.ExpectQuery("SELECT(.+)users").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery("SELECT(.+)policies").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectBegin()
	mock.ExpectExec("INSERT(.+)tasks").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE(.+)downloads").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	asserts.True(monitor.Complete(rpc.StatusInfo{}))
	asserts.NoError(mock.ExpectationsWereMet())
}
