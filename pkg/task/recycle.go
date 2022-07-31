package task

import (
	"context"
	"encoding/json"
	"os"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cluster"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/shadow/slaveinmaster"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

// RecycleTask 文件回收任务
type RecycleTask struct {
	User      *model.User
	TaskModel *model.Task
	TaskProps RecycleProps
	Err       *JobError
}

// RecycleProps 回收任务属性
type RecycleProps struct {
	// 回收目录
	Path string `json:"path"`
	// 负责处理回收任务的节点ID
	NodeID uint `json:"node_id"`
}

// Props 获取任务属性
func (job *RecycleTask) Props() string {
	res, _ := json.Marshal(job.TaskProps)
	return string(res)
}

// Type 获取任务状态
func (job *RecycleTask) Type() int {
	return RecycleTaskType
}

// Creator 获取创建者ID
func (job *RecycleTask) Creator() uint {
	return job.User.ID
}

// Model 获取任务的数据库模型
func (job *RecycleTask) Model() *model.Task {
	return job.TaskModel
}

// SetStatus 设定状态
func (job *RecycleTask) SetStatus(status int) {
	job.TaskModel.SetStatus(status)
}

// SetError 设定任务失败信息
func (job *RecycleTask) SetError(err *JobError) {
	job.Err = err
	res, _ := json.Marshal(job.Err)
	job.TaskModel.SetError(string(res))

}

// SetErrorMsg 设定任务失败信息
func (job *RecycleTask) SetErrorMsg(msg string, err error) {
	jobErr := &JobError{Msg: msg}
	if err != nil {
		jobErr.Error = err.Error()
	}
	job.SetError(jobErr)
}

// GetError 返回任务失败信息
func (job *RecycleTask) GetError() *JobError {
	return job.Err
}

// Do 开始执行任务
func (job *RecycleTask) Do() {
	if job.TaskProps.NodeID == 1 {
		err := os.RemoveAll(job.TaskProps.Path)
		if err != nil {
			util.Log().Warning("无法删除中转临时目录[%s], %s", job.TaskProps.Path, err)
			job.SetErrorMsg("文件回收失败", err)
		}
	} else {
		// 指定为从机回收

		// 创建文件系统
		fs, err := filesystem.NewFileSystem(job.User)
		if err != nil {
			job.SetErrorMsg(err.Error(), nil)
			return
		}

		// 获取从机节点
		node := cluster.Default.GetNodeByID(job.TaskProps.NodeID)
		if node == nil {
			job.SetErrorMsg("从机节点不可用", nil)
		}

		// 切换为从机节点处理回收
		fs.SwitchToSlaveHandler(node)
		handler := fs.Handler.(*slaveinmaster.Driver)
		err = handler.Recycle(context.Background(), job.TaskProps.Path)
		if err != nil {
			util.Log().Warning("无法删除中转临时目录[%s], %s", job.TaskProps.Path, err)
			job.SetErrorMsg("文件回收失败", err)
		}
	}
}

// NewRecycleTask 新建回收任务
func NewRecycleTask(user uint, path string, node uint) (Job, error) {
	creator, err := model.GetActiveUserByID(user)
	if err != nil {
		return nil, err
	}

	newTask := &RecycleTask{
		User: &creator,
		TaskProps: RecycleProps{
			Path:   path,
			NodeID: node,
		},
	}

	record, err := Record(newTask)
	if err != nil {
		return nil, err
	}
	newTask.TaskModel = record

	return newTask, nil
}

// NewRecycleTaskFromModel 从数据库记录中恢复回收任务
func NewRecycleTaskFromModel(task *model.Task) (Job, error) {
	user, err := model.GetActiveUserByID(task.UserID)
	if err != nil {
		return nil, err
	}
	newTask := &RecycleTask{
		User:      &user,
		TaskModel: task,
	}

	err = json.Unmarshal([]byte(task.Props), &newTask.TaskProps)
	if err != nil {
		return nil, err
	}

	return newTask, nil
}
