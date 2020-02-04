package main

import (
	"github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/aria2"
	"github.com/HFO4/cloudreve/pkg/auth"
	"github.com/HFO4/cloudreve/pkg/authn"
	"github.com/HFO4/cloudreve/pkg/cache"
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/HFO4/cloudreve/pkg/task"
	"github.com/HFO4/cloudreve/routers"
	"github.com/gin-gonic/gin"
)

func init() {
	conf.Init("conf/conf.ini")
	// Debug 关闭时，切换为生产模式
	if !conf.SystemConfig.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	cache.Init()
	if conf.SystemConfig.Mode == "master" {
		model.Init()
		authn.Init()
		task.Init()
		aria2.Init()
	}
	auth.Init()
}

func main() {
	api := routers.InitRouter()

	api.Run(conf.SystemConfig.Listen)

}
