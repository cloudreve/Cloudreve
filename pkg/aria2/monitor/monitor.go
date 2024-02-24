package monitor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/common"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/rpc"
	"github.com/cloudreve/Cloudreve/v3/pkg/cluster"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/mq"
	"github.com/cloudreve/Cloudreve/v3/pkg/task"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

// Monitor 离线下载状态监控
type Monitor struct {
	Task     *model.Download
	Interval time.Duration

	notifier <-chan mq.Message
	node     cluster.Node
	retried  int
}

var MAX_RETRY = 10

// NewMonitor 新建离线下载状态监控
func NewMonitor(task *model.Download, pool cluster.Pool, mqClient mq.MQ) {
	monitor := &Monitor{
		Task:     task,
		notifier: make(chan mq.Message),
		node:     pool.GetNodeByID(task.GetNodeID()),
	}

	if monitor.node != nil {
		monitor.Interval = time.Duration(monitor.node.GetAria2Instance().GetConfig().Interval) * time.Second
		go monitor.Loop(mqClient)

		monitor.notifier = mqClient.Subscribe(monitor.Task.GID, 0)
	} else {
		monitor.setErrorStatus(errors.New("node not avaliable"))
	}
}

// Loop 开启监控循环
func (monitor *Monitor) Loop(mqClient mq.MQ) {
	defer mqClient.Unsubscribe(monitor.Task.GID, monitor.notifier)
	fmt.Println(cluster.Default)

	// 首次循环立即更新
	interval := 50 * time.Millisecond

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
	status, err := monitor.node.GetAria2Instance().Status(monitor.Task)

	if err != nil {
		monitor.retried++
		util.Log().Warning("Cannot get status of download task %q: %s", monitor.Task.GID, err)

		// 十次重试后认定为任务失败
		if monitor.retried > MAX_RETRY {
			util.Log().Warning("Cannot get status of download task %q，exceed maximum retry threshold: %s",
				monitor.Task.GID, err)
			monitor.setErrorStatus(err)
			monitor.RemoveTempFolder()
			return true
		}

		return false
	}
	monitor.retried = 0

	// 磁力链下载需要跟随
	if len(status.FollowedBy) > 0 {
		util.Log().Debug("Redirected download task from %q to %q.", monitor.Task.GID, status.FollowedBy[0])
		monitor.Task.GID = status.FollowedBy[0]
		monitor.Task.Save()
		return false
	}

	// 更新任务信息
	if err := monitor.UpdateTaskInfo(status); err != nil {
		util.Log().Warning("Failed to update status of download task %q: %s", monitor.Task.GID, err)
		monitor.setErrorStatus(err)
		monitor.RemoveTempFolder()
		return true
	}

	util.Log().Debug("Remote download %q status updated to %q.", status.Gid, status.Status)

	switch common.GetStatus(status) {
	case common.Complete, common.Seeding:
		return monitor.Complete(task.TaskPoll)
	case common.Error:
		return monitor.Error(status)
	case common.Downloading, common.Ready, common.Paused:
		return false
	case common.Canceled:
		monitor.Task.Status = common.Canceled
		monitor.Task.Save()
		monitor.RemoveTempFolder()
		return true
	default:
		util.Log().Warning("Download task %q returns unknown status %q.", monitor.Task.GID, status.Status)
		return true
	}
}

// UpdateTaskInfo 更新数据库中的任务信息
func (monitor *Monitor) UpdateTaskInfo(status rpc.StatusInfo) error {
	originSize := monitor.Task.TotalSize

	monitor.Task.GID = status.Gid
	monitor.Task.Status = common.GetStatus(status)

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
			monitor.node.GetAria2Instance().Cancel(monitor.Task)
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
		return common.ErrUserNotFound
	}

	// 创建文件系统
	fs, err := filesystem.NewFileSystem(user)
	if err != nil {
		return err
	}
	defer fs.Recycle()

	if err := fs.SetPolicyFromPath(monitor.Task.Dst); err != nil {
		return fmt.Errorf("failed to switch policy to target dir: %w", err)
	}

	// 创建上下文环境
	file := &fsctx.FileStream{
		Size: monitor.Task.TotalSize,
	}

	// 验证用户容量
	if err := filesystem.HookValidateCapacity(context.Background(), fs, file); err != nil {
		return err
	}

	// 验证每个文件
	for _, fileInfo := range monitor.Task.StatusInfo.Files {
		if fileInfo.Selected == "true" {
			// 创建上下文环境
			fileSize, _ := strconv.ParseUint(fileInfo.Length, 10, 64)
			file := &fsctx.FileStream{
				Size: fileSize,
				Name: filepath.Base(fileInfo.Path),
			}
			if err := filesystem.HookValidateFile(context.Background(), fs, file); err != nil {
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
	monitor.node.GetAria2Instance().DeleteTempFile(monitor.Task)
}

// Complete 完成下载，返回是否中断监控
func (monitor *Monitor) Complete(pool task.Pool) bool {
	// 未开始转存，提交转存任务
	if monitor.Task.TaskID == 0 {
		return monitor.transfer(pool)
	}

	// 做种完成
	if common.GetStatus(monitor.Task.StatusInfo) == common.Complete {
		transferTask, err := model.GetTasksByID(monitor.Task.TaskID)
		if err != nil {
			monitor.setErrorStatus(err)
			monitor.RemoveTempFolder()
			return true
		}

		// 转存完成，回收下载目录
		if transferTask.Type == task.TransferTaskType && transferTask.Status >= task.Error {
			job, err := task.NewRecycleTask(monitor.Task)
			if err != nil {
				monitor.setErrorStatus(err)
				monitor.RemoveTempFolder()
				return true
			}

			// 提交回收任务
			pool.Submit(job)

			return true
		}
	}

	return false
}

func (monitor *Monitor) transfer(pool task.Pool) bool {
	// 创建中转任务
	file := make([]string, 0, len(monitor.Task.StatusInfo.Files))
	sizes := make(map[string]uint64, len(monitor.Task.StatusInfo.Files))
	for i := 0; i < len(monitor.Task.StatusInfo.Files); i++ {
		fileInfo := monitor.Task.StatusInfo.Files[i]
		if fileInfo.Selected == "true" {
			file = append(file, fileInfo.Path)
			size, _ := strconv.ParseUint(fileInfo.Length, 10, 64)
			sizes[fileInfo.Path] = size
		}
	}

	job, err := task.NewTransferTask(
		monitor.Task.UserID,
		file,
		monitor.Task.Dst,
		monitor.Task.Parent,
		true,
		monitor.node.ID(),
		sizes,
	)
	if err != nil {
		monitor.setErrorStatus(err)
		monitor.RemoveTempFolder()
		return true
	}

	// 提交中转任务
	pool.Submit(job)

	// 更新任务ID
	monitor.Task.TaskID = job.Model().ID
	monitor.Task.Save()

	return false
}

func (monitor *Monitor) setErrorStatus(err error) {
	monitor.Task.Status = common.Error
	monitor.Task.Error = err.Error()
	monitor.Task.Save()
}
