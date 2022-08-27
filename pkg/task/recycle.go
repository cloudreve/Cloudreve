package task

import (
	"encoding/json"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cluster"
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
	// 下载任务 GID
	DownloadGID string `json:"download_gid"`
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
	download, err := model.GetDownloadByGid(job.TaskProps.DownloadGID, job.User.ID)
	if err != nil {
		util.Log().Warning("回收任务 %d 找不到下载记录", job.TaskModel.ID)
		job.SetErrorMsg("无法找到下载任务", err)
		return
	}
	nodeID := download.GetNodeID()
	node := cluster.Default.GetNodeByID(nodeID)
	if node == nil {
		util.Log().Warning("回收任务 %d 找不到节点", job.TaskModel.ID)
		job.SetErrorMsg("从机节点不可用", nil)
		return
	}
	err = node.GetAria2Instance().DeleteTempFile(download)
	if err != nil {
		util.Log().Warning("无法删除中转临时目录[%s], %s", download.Parent, err)
		job.SetErrorMsg("文件回收失败", err)
		return
	}
}

// NewRecycleTask 新建回收任务
func NewRecycleTask(download *model.Download) (Job, error) {
	newTask := &RecycleTask{
		User: download.GetOwner(),
		TaskProps: RecycleProps{
			DownloadGID: download.GID,
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
