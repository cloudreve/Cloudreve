package bootstrap

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/cluster"
	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/crontab"
	"github.com/cloudreve/Cloudreve/v3/pkg/email"
	"github.com/cloudreve/Cloudreve/v3/pkg/slave"
	"github.com/cloudreve/Cloudreve/v3/pkg/task"
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

	dependencies := []struct {
		mode    string
		factory func()
	}{
		{
			"both",
			func() {
				cache.Init()
			},
		},
		{
			"master",
			func() {
				model.Init()
			},
		},
		{
			"both",
			func() {
				task.Init()
			},
		},
		{
			"master",
			func() {
				cluster.Init()
			},
		},
		{
			"master",
			func() {
				aria2.Init(false)
			},
		},
		{
			"master",
			func() {
				email.Init()
			},
		},
		{
			"master",
			func() {
				crontab.Init()
			},
		},
		{
			"master",
			func() {
				InitStatic()
			},
		},
		{
			"slave",
			func() {
				slave.Init()
			},
		},
		{
			"both",
			func() {
				auth.Init()
			},
		},
	}

	for _, dependency := range dependencies {
		switch dependency.mode {
		case "master":
			if conf.SystemConfig.Mode == "master" {
				dependency.factory()
			}
		case "slave":
			if conf.SystemConfig.Mode == "slave" {
				dependency.factory()
			}
		default:
			dependency.factory()
		}
	}

}
