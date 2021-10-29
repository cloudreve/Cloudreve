package task

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"strings"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cluster"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
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
	Src      []string          `json:"src"`      // 原始文件
	SrcSizes map[string]uint64 `json:"src_size"` // 原始文件的大小信息，从机转存时使用
	Parent   string            `json:"parent"`   // 父目录
	Dst      string            `json:"dst"`      // 目的目录ID
	// 将会保留原始文件的目录结构，Src 除去 Parent 开头作为最终路径
	TrimPath bool `json:"trim_path"`
	// 负责处理中专任务的节点ID
	NodeID uint `json:"node_id"`
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

		dst := path.Join(job.TaskProps.Dst, filepath.Base(file))
		if job.TaskProps.TrimPath {
			// 保留原始目录
			trim := util.FormSlash(job.TaskProps.Parent)
			src := util.FormSlash(file)
			dst = path.Join(job.TaskProps.Dst, strings.TrimPrefix(src, trim))
		}

		ctx := context.WithValue(context.Background(), fsctx.DisableOverwrite, true)
		ctx = context.WithValue(ctx, fsctx.SlaveSrcPath, file)
		if job.TaskProps.NodeID > 1 {
			// 指定为从机中转

			// 获取从机节点
			node := cluster.Default.GetNodeByID(job.TaskProps.NodeID)
			if node == nil {
				job.SetErrorMsg("从机节点不可用", nil)
			}

			// 切换为从机节点处理上传
			fs.SwitchToSlaveHandler(node)
			err = fs.UploadFromStream(ctx, nil, dst, job.TaskProps.SrcSizes[file])
		} else {
			// 主机节点中转
			err = fs.UploadFromPath(ctx, file, dst, true)
		}

		if err != nil {
			job.SetErrorMsg("文件转存失败", err)
		}
	}

}

// Recycle 回收临时文件
func (job *TransferTask) Recycle() {
	if job.TaskProps.NodeID == 1 {
		err := os.RemoveAll(job.TaskProps.Parent)
		if err != nil {
			util.Log().Warning("无法删除中转临时目录[%s], %s", job.TaskProps.Parent, err)
		}
	}
}

// NewTransferTask 新建中转任务
func NewTransferTask(user uint, src []string, dst, parent string, trim bool, node uint, sizes map[string]uint64) (Job, error) {
	creator, err := model.GetActiveUserByID(user)
	if err != nil {
		return nil, err
	}

	newTask := &TransferTask{
		User: &creator,
		TaskProps: TransferProps{
			Src:      src,
			Parent:   parent,
			Dst:      dst,
			TrimPath: trim,
			NodeID:   node,
			SrcSizes: sizes,
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
