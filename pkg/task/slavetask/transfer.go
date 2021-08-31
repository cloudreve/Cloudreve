package slavetask

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/mq"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/slave"
	"github.com/cloudreve/Cloudreve/v3/pkg/task"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"time"
)

// TransferTask 文件中转任务
type TransferTask struct {
	Err      *task.JobError
	Req      *serializer.SlaveTransferReq
	MasterID string
}

// Props 获取任务属性
func (job *TransferTask) Props() string {
	return ""
}

// Type 获取任务类型
func (job *TransferTask) Type() int {
	return 0
}

// Creator 获取创建者ID
func (job *TransferTask) Creator() uint {
	return 0
}

// Model 获取任务的数据库模型
func (job *TransferTask) Model() *model.Task {
	return nil
}

// SetStatus 设定状态
func (job *TransferTask) SetStatus(status int) {
}

// SetError 设定任务失败信息
func (job *TransferTask) SetError(err *task.JobError) {
	job.Err = err

}

// SetErrorMsg 设定任务失败信息
func (job *TransferTask) SetErrorMsg(msg string, err error) {
	jobErr := &task.JobError{Msg: msg}
	if err != nil {
		jobErr.Error = err.Error()
	}
	job.SetError(jobErr)
}

// GetError 返回任务失败信息
func (job *TransferTask) GetError() *task.JobError {
	return job.Err
}

// Do 开始执行任务
func (job *TransferTask) Do() {
	time.Sleep(time.Duration(10) * time.Second)
	msg := mq.Message{
		TriggeredBy: job.MasterID,
		Event:       serializer.SlaveTransferSuccess,
		Content: serializer.SlaveTransferResult{
			Error: nil,
		},
	}

	if err := slave.DefaultController.SendNotification(job.MasterID, job.Req.Hash(job.MasterID), msg); err != nil {
		job.SetErrorMsg("无法发送转存结果通知", err)
	}

	util.Log().Debug("job done")
}
