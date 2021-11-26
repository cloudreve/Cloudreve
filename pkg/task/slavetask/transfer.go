package slavetask

import (
	"context"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cluster"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/mq"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/task"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"os"
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
		util.Log().Warning("无法发送转存失败通知到从机, %s", err)
	}
}

// GetError 返回任务失败信息
func (job *TransferTask) GetError() *task.JobError {
	return job.Err
}

// Do 开始执行任务
func (job *TransferTask) Do() {
	defer job.Recycle()

	fs, err := filesystem.NewAnonymousFileSystem()
	if err != nil {
		job.SetErrorMsg("无法初始化匿名文件系统", err)
		return
	}

	fs.Policy = job.Req.Policy
	if err := fs.DispatchHandler(); err != nil {
		job.SetErrorMsg("无法分发存储策略", err)
		return
	}

	master, err := cluster.DefaultController.GetMasterInfo(job.MasterID)
	if err != nil {
		job.SetErrorMsg("找不到主机节点", err)
		return
	}

	fs.SwitchToShadowHandler(master.Instance, master.URL.String(), master.ID)
	ctx := context.WithValue(context.Background(), fsctx.DisableOverwrite, true)
	file, err := os.Open(util.RelativePath(job.Req.Src))
	if err != nil {
		job.SetErrorMsg("无法读取源文件", err)
		return
	}

	defer file.Close()

	// 获取源文件大小
	fi, err := file.Stat()
	if err != nil {
		job.SetErrorMsg("无法获取源文件大小", err)
		return
	}

	size := fi.Size()

	err = fs.Handler.Put(ctx, file, job.Req.Dst, uint64(size))
	if err != nil {
		job.SetErrorMsg("文件上传失败", err)
		return
	}

	msg := mq.Message{
		TriggeredBy: job.MasterID,
		Event:       serializer.SlaveTransferSuccess,
		Content:     serializer.SlaveTransferResult{},
	}

	if err := cluster.DefaultController.SendNotification(job.MasterID, job.Req.Hash(job.MasterID), msg); err != nil {
		util.Log().Warning("无法发送转存成功通知到从机, %s", err)
	}
}

// Recycle 回收临时文件
func (job *TransferTask) Recycle() {
	err := os.Remove(job.Req.Src)
	if err != nil {
		util.Log().Warning("无法删除中转临时文件[%s], %s", job.Req.Src, err)
	}
}
