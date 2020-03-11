package task

import (
	"context"
	"encoding/json"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem"
	"github.com/HFO4/cloudreve/pkg/util"
	"os"
	"path"
	"path/filepath"
)

// TransferTask 文件中转任务
type TransferTask struct {
	User      *model.User
	TaskModel *model.Task
	TaskProps TransferProps
	Err       *JobError

	zipPath string
}

// TransferProps 中转任务属性
type TransferProps struct {
	Src    []string `json:"src"`    // 原始目录
	Parent string   `json:"parent"` // 父目录
	Dst    string   `json:"dst"`    // 目的目录ID
}

// Props 获取任务属性
func (job *TransferTask) Props() string {
	res, _ := json.Marshal(job.TaskProps)
	return string(res)
}

// Type 获取任务状态
func (job *TransferTask) Type() int {
	return TransferTaskType
}

// Creator 获取创建者ID
func (job *TransferTask) Creator() uint {
	return job.User.ID
}

// Model 获取任务的数据库模型
func (job *TransferTask) Model() *model.Task {
	return job.TaskModel
}

// SetStatus 设定状态
func (job *TransferTask) SetStatus(status int) {
	job.TaskModel.SetStatus(status)
}

// SetError 设定任务失败信息
func (job *TransferTask) SetError(err *JobError) {
	job.Err = err
	res, _ := json.Marshal(job.Err)
	job.TaskModel.SetError(string(res))

}

// SetErrorMsg 设定任务失败信息
func (job *TransferTask) SetErrorMsg(msg string, err error) {
	jobErr := &JobError{Msg: msg}
	if err != nil {
		jobErr.Error = err.Error()
	}
	job.SetError(jobErr)
}

// GetError 返回任务失败信息
func (job *TransferTask) GetError() *JobError {
	return job.Err
}

// Do 开始执行任务
func (job *TransferTask) Do() {
	defer job.Recycle()

	// 创建文件系统
	fs, err := filesystem.NewFileSystem(job.User)
	if err != nil {
		job.SetErrorMsg(err.Error(), nil)
		return
	}

	for index, file := range job.TaskProps.Src {
		job.TaskModel.SetProgress(index)
		err = fs.UploadFromPath(context.Background(), file, path.Join(job.TaskProps.Dst, filepath.Base(file)))
		if err != nil {
			job.SetErrorMsg("文件转存失败", err)
		}
	}

}

// Recycle 回收临时文件
func (job *TransferTask) Recycle() {
	err := os.RemoveAll(job.TaskProps.Parent)
	if err != nil {
		util.Log().Warning("无法删除中转临时目录[%s], %s", job.TaskProps.Parent, err)
	}

}

// NewTransferTask 新建中转任务
func NewTransferTask(user uint, src []string, dst, parent string) (Job, error) {
	creator, err := model.GetActiveUserByID(user)
	if err != nil {
		return nil, err
	}

	newTask := &TransferTask{
		User: &creator,
		TaskProps: TransferProps{
			Src:    src,
			Parent: parent,
			Dst:    dst,
		},
	}

	record, err := Record(newTask)
	if err != nil {
		return nil, err
	}
	newTask.TaskModel = record

	return newTask, nil
}

// NewTransferTaskFromModel 从数据库记录中恢复中转任务
func NewTransferTaskFromModel(task *model.Task) (Job, error) {
	user, err := model.GetActiveUserByID(task.UserID)
	if err != nil {
		return nil, err
	}
	newTask := &TransferTask{
		User:      &user,
		TaskModel: task,
	}

	err = json.Unmarshal([]byte(task.Props), &newTask.TaskProps)
	if err != nil {
		return nil, err
	}

	return newTask, nil
}
