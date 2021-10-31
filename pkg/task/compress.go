package task

import (
	"context"
	"encoding/json"
	"os"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

// CompressTask 文件压缩任务
type CompressTask struct {
	User      *model.User
	TaskModel *model.Task
	TaskProps CompressProps
	Err       *JobError

	zipPath string
}

// CompressProps 压缩任务属性
type CompressProps struct {
	Dirs  []uint `json:"dirs"`
	Files []uint `json:"files"`
	Dst   string `json:"dst"`
}

// Props 获取任务属性
func (job *CompressTask) Props() string {
	res, _ := json.Marshal(job.TaskProps)
	return string(res)
}

// Type 获取任务状态
func (job *CompressTask) Type() int {
	return CompressTaskType
}

// Creator 获取创建者ID
func (job *CompressTask) Creator() uint {
	return job.User.ID
}

// Model 获取任务的数据库模型
func (job *CompressTask) Model() *model.Task {
	return job.TaskModel
}

// SetStatus 设定状态
func (job *CompressTask) SetStatus(status int) {
	job.TaskModel.SetStatus(status)
}

// SetError 设定任务失败信息
func (job *CompressTask) SetError(err *JobError) {
	job.Err = err
	res, _ := json.Marshal(job.Err)
	job.TaskModel.SetError(string(res))

	// 删除压缩文件
	job.removeZipFile()
}

func (job *CompressTask) removeZipFile() {
	if job.zipPath != "" {
		if err := os.Remove(job.zipPath); err != nil {
			util.Log().Warning("无法删除临时压缩文件 %s , %s", job.zipPath, err)
		}
	}
}

// SetErrorMsg 设定任务失败信息
func (job *CompressTask) SetErrorMsg(msg string) {
	job.SetError(&JobError{Msg: msg})
}

// GetError 返回任务失败信息
func (job *CompressTask) GetError() *JobError {
	return job.Err
}

// Do 开始执行任务
func (job *CompressTask) Do() {
	// 创建文件系统
	fs, err := filesystem.NewFileSystem(job.User)
	if err != nil {
		job.SetErrorMsg(err.Error())
		return
	}

	util.Log().Debug("开始压缩文件")
	job.TaskModel.SetProgress(CompressingProgress)

	// 开始压缩
	ctx := context.Background()
	zipFile, err := fs.Compress(ctx, job.TaskProps.Dirs, job.TaskProps.Files, false)
	if err != nil {
		job.SetErrorMsg(err.Error())
		return
	}
	job.zipPath = zipFile

	util.Log().Debug("压缩文件存放至%s，开始上传", zipFile)
	job.TaskModel.SetProgress(TransferringProgress)

	// 上传文件
	err = fs.UploadFromPath(ctx, zipFile, job.TaskProps.Dst, true)
	if err != nil {
		job.SetErrorMsg(err.Error())
		return
	}

	job.removeZipFile()
}

// NewCompressTask 新建压缩任务
func NewCompressTask(user *model.User, dst string, dirs, files []uint) (Job, error) {
	newTask := &CompressTask{
		User: user,
		TaskProps: CompressProps{
			Dirs:  dirs,
			Files: files,
			Dst:   dst,
		},
	}

	record, err := Record(newTask)
	if err != nil {
		return nil, err
	}
	newTask.TaskModel = record

	return newTask, nil
}

// NewCompressTaskFromModel 从数据库记录中恢复压缩任务
func NewCompressTaskFromModel(task *model.Task) (Job, error) {
	user, err := model.GetActiveUserByID(task.UserID)
	if err != nil {
		return nil, err
	}
	newTask := &CompressTask{
		User:      &user,
		TaskModel: task,
	}

	err = json.Unmarshal([]byte(task.Props), &newTask.TaskProps)
	if err != nil {
		return nil, err
	}

	return newTask, nil
}
