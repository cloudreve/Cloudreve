package bootstrap

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/aria2"
	"github.com/HFO4/cloudreve/pkg/auth"
	"github.com/HFO4/cloudreve/pkg/cache"
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/HFO4/cloudreve/pkg/crontab"
	"github.com/HFO4/cloudreve/pkg/email"
	"github.com/HFO4/cloudreve/pkg/task"
	"github.com/gin-gonic/gin"
)

// Init 初始化启动
func Init(path string) {
	InitApplication()
	conf.Init(path)
	// Debug 关闭时，切换为生产模式
	if !conf.SystemConfig.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	cache.Init()
	if conf.SystemConfig.Mode == "master" {
		model.Init()
		task.Init()
		aria2.Init(false)
		email.Init()
		crontab.Init()
		InitStatic()
	}
	auth.Init()
}
