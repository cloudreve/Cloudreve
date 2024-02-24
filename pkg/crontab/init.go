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
	util.Log().Info("Initialize crontab jobs...")
	// 读取cron日程设置
	options := model.GetSettingByNames(
		"cron_garbage_collect",
		"cron_notify_user",
		"cron_ban_user",
		"cron_recycle_upload_session",
	)
	Cron := cron.New()
	for k, v := range options {
		var handler func()
		switch k {
		case "cron_garbage_collect":
			handler = garbageCollect
		case "cron_notify_user":
			handler = notifyExpiredVAS
		case "cron_ban_user":
			handler = banOverusedUser
		case "cron_recycle_upload_session":
			handler = uploadSessionCollect
		default:
			util.Log().Warning("Unknown crontab job type %q, skipping...", k)
			continue
		}

		if _, err := Cron.AddFunc(v, handler); err != nil {
			util.Log().Warning("Failed to start crontab job %q: %s", k, err)
		}

	}
	Cron.Start()
}
