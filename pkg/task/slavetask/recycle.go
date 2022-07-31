package slavetask

import (
	"os"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cluster"
	"github.com/cloudreve/Cloudreve/v3/pkg/mq"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/task"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

// RecycleTask 文件回收任务
type RecycleTask struct {
	Err      *task.JobError
	Req      *serializer.SlaveRecycleReq
	MasterID string
}

// Props 获取任务属性
func (job *RecycleTask) Props() string {
	return ""
}

// Type 获取任务类型
func (job *RecycleTask) Type() int {
	return 0
}

// Creator 获取创建者ID
func (job *RecycleTask) Creator() uint {
	return 0
}

// Model 获取任务的数据库模型
func (job *RecycleTask) Model() *model.Task {
	return nil
}

// SetStatus 设定状态
func (job *RecycleTask) SetStatus(status int) {
}

// SetError 设定任务失败信息
func (job *RecycleTask) SetError(err *task.JobError) {
	job.Err = err
}

// SetErrorMsg 设定任务失败信息
func (job *RecycleTask) SetErrorMsg(msg string, err error) {
	jobErr := &task.JobError{Msg: msg}
	if err != nil {
		jobErr.Error = err.Error()
	}

	job.SetError(jobErr)

	notifyMsg := mq.Message{
		TriggeredBy: job.MasterID,
		Event:       serializer.SlaveRecycleFailed,
		Content: serializer.SlaveRecycleResult{
			Error: err.Error(),
		},
	}

	if err = cluster.DefaultController.SendNotification(job.MasterID, job.Req.Hash(job.MasterID), notifyMsg); err != nil {
		util.Log().Warning("无法发送回收失败通知到从机, %s", err)
	}
}

// GetError 返回任务失败信息
func (job *RecycleTask) GetError() *task.JobError {
	return job.Err
}

// Do 开始执行任务
func (job *RecycleTask) Do() {
	err := os.RemoveAll(job.Req.Path)
	if err != nil {
		util.Log().Warning("无法删除中转临时文件[%s], %s", job.Req.Path, err)
		job.SetErrorMsg("文件回收失败", err)
		return
	}

	msg := mq.Message{
		TriggeredBy: job.MasterID,
		Event:       serializer.SlaveRecycleSuccess,
		Content:     serializer.SlaveRecycleResult{},
	}

	if err = cluster.DefaultController.SendNotification(job.MasterID, job.Req.Hash(job.MasterID), msg); err != nil {
		util.Log().Warning("无法发送回收成功通知到从机, %s", err)
	}
}
