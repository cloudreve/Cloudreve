package task

import (
	"fmt"

	"github.com/cloudreve/Cloudreve/v3/pkg/logger"
)

// Worker 处理任务的对象
type Worker interface {
	Do(Job) // 执行任务
}

// GeneralWorker 通用Worker
type GeneralWorker struct {
}

// Do 执行任务
func (worker *GeneralWorker) Do(job Job) {
	logger.Debug("开始执行任务")
	job.SetStatus(Processing)

	defer func() {
		// 致命错误捕获
		if err := recover(); err != nil {
			logger.Debug("任务执行出错，%s", err)
			job.SetError(&JobError{Msg: "致命错误", Error: fmt.Sprintf("%s", err)})
			job.SetStatus(Error)
		}
	}()

	// 开始执行任务
	job.Do()

	// 任务执行失败
	if err := job.GetError(); err != nil {
		logger.Debug("任务执行出错")
		job.SetStatus(Error)
		return
	}

	logger.Debug("任务执行完成")
	// 执行完成
	job.SetStatus(Complete)
}
