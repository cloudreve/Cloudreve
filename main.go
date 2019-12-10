package main

import (
	"github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/auth"
	"github.com/HFO4/cloudreve/pkg/authn"
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/HFO4/cloudreve/routers"
	"github.com/gin-gonic/gin"
)

func init() {
	conf.Init("conf/conf.ini")
	model.Init()

	// Debug 关闭时，切换为生产模式
	if !conf.SystemConfig.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	auth.Init()
	authn.Init()
}

func main() {
	api := routers.InitRouter()

	api.Run(":5000")

}
