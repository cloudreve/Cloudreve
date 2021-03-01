package task

import (
	"context"
	"encoding/json"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
)

// DecompressTask 文件压缩任务
type DecompressTask struct {
	User      *model.User
	TaskModel *model.Task
	TaskProps DecompressProps
	Err       *JobError

	zipPath string
}

// DecompressProps 压缩任务属性
type DecompressProps struct {
	Src string `json:"src"`
	Dst string `json:"dst"`
}

// Props 获取任务属性
func (job *DecompressTask) Props() string {
	res, _ := json.Marshal(job.TaskProps)
	return string(res)
}

// Type 获取任务状态
func (job *DecompressTask) Type() int {
	return DecompressTaskType
}

// Creator 获取创建者ID
func (job *DecompressTask) Creator() uint {
	return job.User.ID
}

// Model 获取任务的数据库模型
func (job *DecompressTask) Model() *model.Task {
	return job.TaskModel
}

// SetStatus 设定状态
func (job *DecompressTask) SetStatus(status int) {
	job.TaskModel.SetStatus(status)
}

// SetError 设定任务失败信息
func (job *DecompressTask) SetError(err *JobError) {
	job.Err = err
	res, _ := json.Marshal(job.Err)
	job.TaskModel.SetError(string(res))
}

// SetErrorMsg 设定任务失败信息
func (job *DecompressTask) SetErrorMsg(msg string, err error) {
	jobErr := &JobError{Msg: msg}
	if err != nil {
		jobErr.Error = err.Error()
	}
	job.SetError(jobErr)
}

// GetError 返回任务失败信息
func (job *DecompressTask) GetError() *JobError {
	return job.Err
}

// Do 开始执行任务
func (job *DecompressTask) Do() {
	// 创建文件系统
	fs, err := filesystem.NewFileSystem(job.User)
	if err != nil {
		job.SetErrorMsg("无法创建文件系统", err)
		return
	}

	job.TaskModel.SetProgress(DecompressingProgress)

	// 禁止重名覆盖
	ctx := context.Background()
	ctx = context.WithValue(ctx, fsctx.DisableOverwrite, true)

	err = fs.Decompress(ctx, job.TaskProps.Src, job.TaskProps.Dst)
	if err != nil {
		job.SetErrorMsg("解压缩失败", err)
		return
	}

}

// NewDecompressTask 新建压缩任务
func NewDecompressTask(user *model.User, src, dst string) (Job, error) {
	newTask := &DecompressTask{
		User: user,
		TaskProps: DecompressProps{
			Src: src,
			Dst: dst,
		},
	}

	record, err := Record(newTask)
	if err != nil {
		return nil, err
	}
	newTask.TaskModel = record

	return newTask, nil
}

// NewDecompressTaskFromModel 从数据库记录中恢复压缩任务
func NewDecompressTaskFromModel(task *model.Task) (Job, error) {
	user, err := model.GetActiveUserByID(task.UserID)
	if err != nil {
		return nil, err
	}
	newTask := &DecompressTask{
		User:      &user,
		TaskModel: task,
	}

	err = json.Unmarshal([]byte(task.Props), &newTask.TaskProps)
	if err != nil {
		return nil, err
	}

	return newTask, nil
}
