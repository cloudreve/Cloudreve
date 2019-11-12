package routers

import (
	"cloudreve/middleware"
	"cloudreve/pkg/conf"
	"cloudreve/routers/controllers"
	"github.com/gin-gonic/gin"
)

// InitRouter 初始化路由
func InitRouter() *gin.Engine {
	r := gin.Default()

	// 中间件
	r.Use(middleware.Session(conf.SystemConfig.SessionSecret))
	r.Use(middleware.CurrentUser())

	// 顶层路由分组
	v3 := r.Group("/Api/V3")
	{
		// 测试用路由
		v3.GET("Ping", controllers.Ping)
		// 用户登录
		v3.POST("User/Session", controllers.UserLogin)

		// 需要登录保护的
		auth := v3.Group("")
		auth.Use(middleware.AuthRequired())
		{
			// 用户类
			user := auth.Group("User")
			{
				// 当前登录用户信息
				user.GET("Me", controllers.UserMe)
			}

		}

	}
	return r
}
