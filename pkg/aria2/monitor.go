package aria2

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/rpc"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/local"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/task"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

// Monitor 离线下载状态监控
type Monitor struct {
	Task     *model.Download
	Interval time.Duration

	notifier chan StatusEvent
	retried  int
}

// StatusEvent 状态改变事件
type StatusEvent struct {
	GID    string
	Status int
}

var MAX_RETRY = 10

// NewMonitor 新建上传状态监控
func NewMonitor(task *model.Download) {
	monitor := &Monitor{
		Task:     task,
		Interval: time.Duration(model.GetIntSetting("aria2_interval", 10)) * time.Second,
		notifier: make(chan StatusEvent),
	}
	go monitor.Loop()
	EventNotifier.Subscribe(monitor.notifier, monitor.Task.GID)
}

// Loop 开启监控循环
func (monitor *Monitor) Loop() {
	defer EventNotifier.Unsubscribe(monitor.Task.GID)

	// 首次循环立即更新
	interval := time.Duration(0)

	for {
		select {
		case <-monitor.notifier:
			if monitor.Update() {
				return
			}
		case <-time.After(interval):
			interval = monitor.Interval
			if monitor.Update() {
				return
			}
		}
	}
}

// Update 更新状态，返回值表示是否退出监控
func (monitor *Monitor) Update() bool {
	Lock.RLock()
	status, err := Instance.Status(monitor.Task)
	Lock.RUnlock()

	if err != nil {
		monitor.retried++
		util.Log().Warning("无法获取下载任务[%s]的状态，%s", monitor.Task.GID, err)

		// 十次重试后认定为任务失败
		if monitor.retried > MAX_RETRY {
			util.Log().Warning("无法获取下载任务[%s]的状态，超过最大重试次数限制，%s", monitor.Task.GID, err)
			monitor.setErrorStatus(err)
			monitor.RemoveTempFolder()
			return true
		}

		return false
	}
	monitor.retried = 0

	// 磁力链下载需要跟随
	if len(status.FollowedBy) > 0 {
		util.Log().Debug("离线下载[%s]重定向至[%s]", monitor.Task.GID, status.FollowedBy[0])
		monitor.Task.GID = status.FollowedBy[0]
		monitor.Task.Save()
		return false
	}

	// 更新任务信息
	if err := monitor.UpdateTaskInfo(status); err != nil {
		util.Log().Warning("无法更新下载任务[%s]的任务信息[%s]，", monitor.Task.GID, err)
		monitor.setErrorStatus(err)
		return true
	}

	util.Log().Debug("离线下载[%s]更新状态[%s]", status.Gid, status.Status)

	switch status.Status {
	case "complete":
		return monitor.Complete(status)
	case "error":
		return monitor.Error(status)
	case "active", "waiting", "paused":
		return false
	case "removed":
		monitor.Task.Status = Canceled
		monitor.Task.Save()
		monitor.RemoveTempFolder()
		return true
	default:
		util.Log().Warning("下载任务[%s]返回未知状态信息[%s]，", monitor.Task.GID, status.Status)
		return true
	}
}

// UpdateTaskInfo 更新数据库中的任务信息
func (monitor *Monitor) UpdateTaskInfo(status rpc.StatusInfo) error {
	originSize := monitor.Task.TotalSize

	monitor.Task.GID = status.Gid
	monitor.Task.Status = getStatus(status.Status)

	// 文件大小、已下载大小
	total, err := strconv.ParseUint(status.TotalLength, 10, 64)
	if err != nil {
		total = 0
	}
	downloaded, err := strconv.ParseUint(status.CompletedLength, 10, 64)
	if err != nil {
		downloaded = 0
	}
	monitor.Task.TotalSize = total
	monitor.Task.DownloadedSize = downloaded
	monitor.Task.GID = status.Gid
	monitor.Task.Parent = status.Dir

	// 下载速度
	speed, err := strconv.Atoi(status.DownloadSpeed)
	if err != nil {
		speed = 0
	}

	monitor.Task.Speed = speed
	attrs, _ := json.Marshal(status)
	monitor.Task.Attrs = string(attrs)

	if err := monitor.Task.Save(); err != nil {
		return err
	}

	if originSize != monitor.Task.TotalSize {
		// 文件大小更新后，对文件限制等进行校验
		if err := monitor.ValidateFile(); err != nil {
			// 验证失败时取消任务
			Lock.RLock()
			Instance.Cancel(monitor.Task)
			Lock.RUnlock()
			return err
		}
	}

	return nil
}

// ValidateFile 上传过程中校验文件大小、文件名
func (monitor *Monitor) ValidateFile() error {
	// 找到任务创建者
	user := monitor.Task.GetOwner()
	if user == nil {
		return ErrUserNotFound
	}

	// 创建文件系统
	fs, err := filesystem.NewFileSystem(user)
	if err != nil {
		return err
	}
	defer fs.Recycle()

	// 创建上下文环境
	ctx := context.WithValue(context.Background(), fsctx.FileHeaderCtx, local.FileStream{
		Size: monitor.Task.TotalSize,
	})

	// 验证用户容量
	if err := filesystem.HookValidateCapacityWithoutIncrease(ctx, fs); err != nil {
		return err
	}

	// 验证每个文件
	for _, fileInfo := range monitor.Task.StatusInfo.Files {
		if fileInfo.Selected == "true" {
			// 创建上下文环境
			fileSize, _ := strconv.ParseUint(fileInfo.Length, 10, 64)
			ctx := context.WithValue(context.Background(), fsctx.FileHeaderCtx, local.FileStream{
				Size: fileSize,
				Name: filepath.Base(fileInfo.Path),
			})
			if err := filesystem.HookValidateFile(ctx, fs); err != nil {
				return err
			}
		}

	}

	return nil
}

// Error 任务下载出错处理，返回是否中断监控
func (monitor *Monitor) Error(status rpc.StatusInfo) bool {
	monitor.setErrorStatus(errors.New(status.ErrorMessage))

	// 清理临时文件
	monitor.RemoveTempFolder()

	return true
}

// RemoveTempFolder 清理下载临时目录
func (monitor *Monitor) RemoveTempFolder() {
	err := os.RemoveAll(monitor.Task.Parent)
	if err != nil {
		util.Log().Warning("无法删除离线下载临时目录[%s], %s", monitor.Task.Parent, err)
	}

}

// Complete 完成下载，返回是否中断监控
func (monitor *Monitor) Complete(status rpc.StatusInfo) bool {
	// 创建中转任务
	file := make([]string, 0, len(monitor.Task.StatusInfo.Files))
	for i := 0; i < len(monitor.Task.StatusInfo.Files); i++ {
		if monitor.Task.StatusInfo.Files[i].Selected == "true" {
			file = append(file, monitor.Task.StatusInfo.Files[i].Path)
		}
	}
	job, err := task.NewTransferTask(
		monitor.Task.UserID,
		file,
		monitor.Task.Dst,
		monitor.Task.Parent,
		true,
	)
	if err != nil {
		monitor.setErrorStatus(err)
		return true
	}

	// 提交中转任务
	task.TaskPoll.Submit(job)

	// 更新任务ID
	monitor.Task.TaskID = job.Model().ID
	monitor.Task.Save()

	return true
}

func (monitor *Monitor) setErrorStatus(err error) {
	monitor.Task.Status = Error
	monitor.Task.Error = err.Error()
	monitor.Task.Save()
}
