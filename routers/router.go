package routers

import (
	"github.com/HFO4/cloudreve/middleware"
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/HFO4/cloudreve/routers/controllers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// InitRouter 初始化路由
func InitRouter() *gin.Engine {
	r := gin.Default()

	/*
		中间件
	*/
	r.Use(middleware.Session(conf.SystemConfig.SessionSecret))

	// CORS TODO: 根据配置文件来
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"PUT", "POST", "GET", "OPTIONS"},
		AllowHeaders:     []string{"X-PINGOTHER", "Content-Type"},
		AllowCredentials: true,
	}))

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
		v3.GET("Site/Ping", controllers.Ping)
		// 用户登录
		v3.POST("User/Session", controllers.UserLogin)
		// 验证码
		v3.GET("Captcha", controllers.Captcha)
		// 站点全局配置
		v3.GET("Site/Config", controllers.SiteConfig)

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
				// 文件上传
				file.POST("Upload", controllers.FileUploadStream)
			}

		}

	}
	return r
}
