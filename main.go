package main

import (
	"cloudreve/models"
	"cloudreve/pkg/conf"
	"cloudreve/routers"
	"github.com/gin-gonic/gin"
	"math/rand"
	"time"
)

func init() {
	conf.Init("conf/conf.ini")
	model.Init()

	rand.Seed(time.Now().UnixNano())

	// Debug 关闭时，切换为生产模式
	if !conf.SystemConfig.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
}

func main() {
	api := routers.InitRouter()

	api.Run(":5000")

}
