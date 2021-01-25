package crontab

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/robfig/cron/v3"
)

// Cron 定时任务
var Cron *cron.Cron

// Reload 重新启动定时任务
func Reload() {
	if Cron != nil {
		Cron.Stop()
	}
	Init()
}

// Init 初始化定时任务
func Init() {
	util.Log().Info("初始化定时任务...")
	// 读取cron日程设置
	options := model.GetSettingByNames("cron_garbage_collect")
	Cron := cron.New()
	for k, v := range options {
		var handler func()
		switch k {
		case "cron_garbage_collect":
			handler = garbageCollect
		default:
			util.Log().Warning("未知定时任务类型 [%s]，跳过", k)
			continue
		}

		if _, err := Cron.AddFunc(v, handler); err != nil {
			util.Log().Warning("无法启动定时任务 [%s] , %s", k, err)
		}

	}
	Cron.Start()
}
