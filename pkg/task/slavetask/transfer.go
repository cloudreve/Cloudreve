package slavetask

import (
	"context"
	"os"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cluster"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/mq"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/task"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
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

	notifyMsg := mq.Message{
		TriggeredBy: job.MasterID,
		Event:       serializer.SlaveTransferFailed,
		Content: serializer.SlaveTransferResult{
			Error: err.Error(),
		},
	}

	if err := cluster.DefaultController.SendNotification(job.MasterID, job.Req.Hash(job.MasterID), notifyMsg); err != nil {
		util.Log().Warning("Failed to send transfer failure notification to master node: %s", err)
	}
}

// GetError 返回任务失败信息
func (job *TransferTask) GetError() *task.JobError {
	return job.Err
}

// Do 开始执行任务
func (job *TransferTask) Do() {
	fs, err := filesystem.NewAnonymousFileSystem()
	if err != nil {
		job.SetErrorMsg("Failed to initialize anonymous filesystem.", err)
		return
	}

	fs.Policy = job.Req.Policy
	if err := fs.DispatchHandler(); err != nil {
		job.SetErrorMsg("Failed to dispatch policy.", err)
		return
	}

	master, err := cluster.DefaultController.GetMasterInfo(job.MasterID)
	if err != nil {
		job.SetErrorMsg("Cannot found master node ID.", err)
		return
	}

	fs.SwitchToShadowHandler(master.Instance, master.URL.String(), master.ID)
	file, err := os.Open(util.RelativePath(job.Req.Src))
	if err != nil {
		job.SetErrorMsg("Failed to read source file.", err)
		return
	}

	defer file.Close()

	// 获取源文件大小
	fi, err := file.Stat()
	if err != nil {
		job.SetErrorMsg("Failed to get source file size.", err)
		return
	}

	size := fi.Size()

	err = fs.Handler.Put(context.Background(), &fsctx.FileStream{
		File:     file,
		SavePath: job.Req.Dst,
		Size:     uint64(size),
	})
	if err != nil {
		job.SetErrorMsg("Upload failed.", err)
		return
	}

	msg := mq.Message{
		TriggeredBy: job.MasterID,
		Event:       serializer.SlaveTransferSuccess,
		Content:     serializer.SlaveTransferResult{},
	}

	if err := cluster.DefaultController.SendNotification(job.MasterID, job.Req.Hash(job.MasterID), msg); err != nil {
		util.Log().Warning("Failed to send transfer success notification to master node: %s", err)
	}
}
