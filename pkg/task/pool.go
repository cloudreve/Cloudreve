package task

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

// TaskPoll 要使用的任务池
var TaskPoll Pool

type Pool interface {
	Add(num int)
	Submit(job Job)
}

// AsyncPool 带有最大配额的任务池
type AsyncPool struct {
	// 容量
	idleWorker chan int
}

// Add 增加可用Worker数量
func (pool *AsyncPool) Add(num int) {
	for i := 0; i < num; i++ {
		pool.idleWorker <- 1
	}
}

// ObtainWorker 阻塞直到获取新的Worker
func (pool *AsyncPool) obtainWorker() Worker {
	select {
	case <-pool.idleWorker:
		// 有空闲Worker名额时，返回新Worker
		return &GeneralWorker{}
	}
}

// FreeWorker 添加空闲Worker
func (pool *AsyncPool) freeWorker() {
	pool.Add(1)
}

// Submit 开始提交任务
func (pool *AsyncPool) Submit(job Job) {
	go func() {
		util.Log().Debug("等待获取Worker")
		worker := pool.obtainWorker()
		util.Log().Debug("获取到Worker")
		worker.Do(job)
		util.Log().Debug("释放Worker")
		pool.freeWorker()
	}()
}

// Init 初始化任务池
func Init() {
	maxWorker := model.GetIntSetting("max_worker_num", 10)
	TaskPoll = &AsyncPool{
		idleWorker: make(chan int, maxWorker),
	}
	TaskPoll.Add(maxWorker)
	util.Log().Info("初始化任务队列，WorkerNum = %d", maxWorker)

	if conf.SystemConfig.Mode == "master" {
		Resume(TaskPoll)
	}
}
