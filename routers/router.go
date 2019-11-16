package routers

import (
	"github.com/HFO4/cloudreve/middleware"
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/HFO4/cloudreve/routers/controllers"
	"github.com/gin-gonic/gin"
)

// InitRouter 初始化路由
func InitRouter() *gin.Engine {
	r := gin.Default()

	/*
		中间件
	*/
	r.Use(middleware.Session(conf.SystemConfig.SessionSecret))

	// 测试模式加入Mock助手中间件
	if gin.Mode() == gin.TestMode {
		r.Use(middleware.MockHelper())
	}

	r.Use(middleware.CurrentUser())

	/*
		路由
	*/
	v3 := r.Group("/Api/V3")
	{
		// 测试用路由
		v3.GET("Ping", controllers.Ping)
		// 用户登录
		v3.POST("User/Session", controllers.UserLogin)
		// 验证码
		v3.GET("Captcha", controllers.Captcha)

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

			// 文件
			file := auth.Group("File")
			{
				// 当前登录用户信息
				file.POST("Upload", controllers.FileUpload)
			}

		}

	}
	return r
}
