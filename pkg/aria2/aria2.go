package aria2

import (
	"sync"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/common"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/monitor"
	"github.com/cloudreve/Cloudreve/v3/pkg/balancer"
)

// Instance 默认使用的Aria2处理实例
var Instance common.Aria2 = &common.DummyAria2{}

// LB 获取 Aria2 节点的负载均衡器
var LB balancer.Balancer

// Lock Instance的读写锁
var Lock sync.RWMutex

// Init 初始化
func Init(isReload bool) {
	Lock.Lock()
	LB = balancer.NewBalancer("RoundRobin")
	Lock.Unlock()

	if !isReload {
		// 从数据库中读取未完成任务，创建监控
		unfinished := model.GetDownloadsByStatus(common.Ready, common.Paused, common.Downloading)

		for i := 0; i < len(unfinished); i++ {
			// 创建任务监控
			monitor.NewMonitor(&unfinished[i])
		}
	}
}
