package task

import (
	"context"
	"encoding/json"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

// RelocateTask 存储策略迁移任务
type RelocateTask struct {
	User      *model.User
	TaskModel *model.Task
	TaskProps RelocateProps
	Err       *JobError
}

// RelocateProps 存储策略迁移任务属性
type RelocateProps struct {
	Dirs        []uint `json:"dirs"`
	Files       []uint `json:"files"`
	DstPolicyID uint   `json:"dst_policy_id"`
}

// Props 获取任务属性
func (job *RelocateTask) Props() string {
	res, _ := json.Marshal(job.TaskProps)
	return string(res)
}

// Type 获取任务状态
func (job *RelocateTask) Type() int {
	return RelocateTaskType
}

// Creator 获取创建者ID
func (job *RelocateTask) Creator() uint {
	return job.User.ID
}

// Model 获取任务的数据库模型
func (job *RelocateTask) Model() *model.Task {
	return job.TaskModel
}

// SetStatus 设定状态
func (job *RelocateTask) SetStatus(status int) {
	job.TaskModel.SetStatus(status)
}

// SetError 设定任务失败信息
func (job *RelocateTask) SetError(err *JobError) {
	job.Err = err
	res, _ := json.Marshal(job.Err)
	job.TaskModel.SetError(string(res))
}

// SetErrorMsg 设定任务失败信息
func (job *RelocateTask) SetErrorMsg(msg string) {
	job.SetError(&JobError{Msg: msg})
}

// GetError 返回任务失败信息
func (job *RelocateTask) GetError() *JobError {
	return job.Err
}

// Do 开始执行任务
func (job *RelocateTask) Do() {
	// 创建文件系统
	fs, err := filesystem.NewFileSystem(job.User)
	if err != nil {
		job.SetErrorMsg(err.Error())
		return
	}

	job.TaskModel.SetProgress(ListingProgress)
	util.Log().Debug("Start migration task.")

	// ----------------------------
	//   索引出所有待迁移的文件
	// ----------------------------
	targetFiles := make([]model.File, 0, len(job.TaskProps.Files))

	// 索引用户选择的单独的文件
	outerFiles, err := model.GetFilesByIDs(job.TaskProps.Files, job.User.ID)
	if err != nil {
		job.SetError(&JobError{
			Msg:   "Failed to index files.",
			Error: err.Error(),
		})
		return
	}
	targetFiles = append(targetFiles, outerFiles...)

	// 索引用户选择目录下的所有递归子文件
	subFolders, err := model.GetRecursiveChildFolder(job.TaskProps.Dirs, job.User.ID, true)
	if err != nil {
		job.SetError(&JobError{
			Msg:   "Failed to index child folders.",
			Error: err.Error(),
		})
		return
	}

	subFiles, err := model.GetChildFilesOfFolders(&subFolders)
	if err != nil {
		job.SetError(&JobError{
			Msg:   "Failed to index child files.",
			Error: err.Error(),
		})
		return
	}
	targetFiles = append(targetFiles, subFiles...)

	// 查找目标存储策略
	policy, err := model.GetPolicyByID(job.TaskProps.DstPolicyID)
	if err != nil {
		job.SetError(&JobError{
			Msg:   "Invalid policy.",
			Error: err.Error(),
		})
		return
	}

	// 开始转移文件
	job.TaskModel.SetProgress(TransferringProgress)
	ctx := context.Background()
	err = fs.Relocate(ctx, targetFiles, &policy)
	if err != nil {
		job.SetErrorMsg(err.Error())
		return
	}

	return
}

// NewRelocateTask 新建转移任务
func NewRelocateTask(user *model.User, dstPolicyID uint, dirs, files []uint) (Job, error) {
	newTask := &RelocateTask{
		User: user,
		TaskProps: RelocateProps{
			Dirs:        dirs,
			Files:       files,
			DstPolicyID: dstPolicyID,
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
func NewRelocateTaskFromModel(task *model.Task) (Job, error) {
	user, err := model.GetActiveUserByID(task.UserID)
	if err != nil {
		return nil, err
	}
	newTask := &RelocateTask{
		User:      &user,
		TaskModel: task,
	}

	err = json.Unmarshal([]byte(task.Props), &newTask.TaskProps)
	if err != nil {
		return nil, err
	}

	return newTask, nil
}
