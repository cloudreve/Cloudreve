package routers

import (
	"cloudreve/middleware"
	"cloudreve/pkg/conf"
	"cloudreve/routers/controllers"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()

	// 中间件
	r.Use(middleware.Session(conf.SystemConfig.SessionSecret))

	// 顶层路由分组
	v3 := r.Group("/Api/V3")
	{
		// 测试用路由
		v3.GET("Ping", controllers.Ping)
		// 用户登录
		v3.POST("User/Session", controllers.UserLogin)

	}
	return r
}
