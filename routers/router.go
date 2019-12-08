package routers

import (
	"github.com/HFO4/cloudreve/middleware"
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/HFO4/cloudreve/routers/controllers"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

// InitRouter 初始化路由
func InitRouter() *gin.Engine {
	r := gin.Default()
	pprof.Register(r)

	/*
		中间件
	*/
	r.Use(middleware.Session(conf.SystemConfig.SessionSecret))

	// CORS TODO: 根据配置文件来
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"PUT", "POST", "GET", "OPTIONS"},
		AllowHeaders:     []string{"Cookie", "Content-Length", "Content-Type", "X-Path", "X-FileName"},
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
	v3 := r.Group("/api/v3")
	{
		// 测试用路由
		v3.GET("site/ping", controllers.Ping)
		// 用户登录
		v3.POST("user/session", controllers.UserLogin)
		// WebAuthn登陆初始化
		v3.GET("user/authn/:username", controllers.StartLoginAuthn)
		// WebAuthn登陆
		v3.POST("user/authn/finish/:username", controllers.FinishLoginAuthn)
		// 验证码
		v3.GET("captcha", controllers.Captcha)
		// 站点全局配置
		v3.GET("site/config", controllers.SiteConfig)

		// 需要登录保护的
		auth := v3.Group("")
		auth.Use(middleware.AuthRequired())
		{
			// 用户
			user := auth.Group("user")
			{
				// 当前登录用户信息
				user.GET("me", controllers.UserMe)
				user.GET("storage", controllers.UserStorage)

				// WebAuthn 注册相关
				authn := user.Group("authn")
				{
					authn.PUT("", controllers.StartRegAuthn)
					authn.PUT("finish", controllers.FinishRegAuthn)
				}
			}

			// 文件
			file := auth.Group("file")
			{
				// 文件上传
				file.POST("upload", controllers.FileUploadStream)
				// 下载文件
				file.GET("download/*path", controllers.Download)
				// 下载文件
				file.GET("thumb/:id", controllers.Thumb)
			}

			// 目录
			directory := auth.Group("directory")
			{
				// 创建目录
				directory.PUT("", controllers.CreateDirectory)
				// 列出目录下内容
				directory.GET("*path", controllers.ListDirectory)
			}

			// 对象，文件和目录的抽象
			object := auth.Group("object")
			{
				// 删除对象
				object.DELETE("", controllers.Delete)
				// 移动对象
				object.PATCH("", controllers.Move)
				// 复制对象
				object.POST("copy", controllers.Copy)
				// 重命名对象
				object.POST("rename", controllers.Rename)
			}

		}

	}
	return r
}
