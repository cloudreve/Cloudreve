package routers

import (
	"Cloudreve/routers/controllers"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()

	// 路由
	v3 := r.Group("/Api/V3")
	{
		v3.GET("Ping", controllers.Ping)

	}
	return r
}
