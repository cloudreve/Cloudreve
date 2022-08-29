package crontab

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/logger"
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
	logger.Info("初始化定时任务...")
	// 读取cron日程设置
	options := model.GetSettingByNames(
		"cron_garbage_collect",
		"cron_recycle_upload_session",
	)
	Cron := cron.New()
	for k, v := range options {
		var handler func()
		switch k {
		case "cron_garbage_collect":
			handler = garbageCollect
		case "cron_recycle_upload_session":
			handler = uploadSessionCollect
		default:
			logger.Warning("未知定时任务类型 [%s]，跳过", k)
			continue
		}

		if _, err := Cron.AddFunc(v, handler); err != nil {
			logger.Warning("无法启动定时任务 [%s] , %s", k, err)
		}

	}
	Cron.Start()
}
