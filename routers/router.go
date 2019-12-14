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
	if conf.CORSConfig.AllowOrigins[0] != "UNSET" || conf.CORSConfig.AllowAllOrigins {
		r.Use(cors.New(cors.Config{
			AllowOrigins:     conf.CORSConfig.AllowOrigins,
			AllowAllOrigins:  conf.CORSConfig.AllowAllOrigins,
			AllowMethods:     conf.CORSConfig.AllowHeaders,
			AllowHeaders:     conf.CORSConfig.AllowHeaders,
			AllowCredentials: conf.CORSConfig.AllowCredentials,
			ExposeHeaders:    conf.CORSConfig.ExposeHeaders,
		}))
	}

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
		// 全局设置相关
		site := v3.Group("site")
		{
			// 测试用路由
			site.GET("ping", controllers.Ping)
			// 验证码
			site.GET("captcha", controllers.Captcha)
			// 站点全局配置
			site.GET("config", controllers.SiteConfig)
		}

		// 用户相关路由
		user := v3.Group("user")
		{
			// 用户登录
			user.POST("session", controllers.UserLogin)
			// WebAuthn登陆初始化
			user.GET("authn/:username", controllers.StartLoginAuthn)
			// WebAuthn登陆
			user.POST("authn/finish/:username", controllers.FinishLoginAuthn)
		}

		// 需要携带签名验证的
		sign := v3.Group("")
		sign.Use(middleware.SignRequired())
		{
			file := sign.Group("file")
			{
				// 文件外链
				file.GET("get/:id/:name", controllers.AnonymousGetContent)
				// 下載已经打包好的文件
				file.GET("archive/:id/archive.zip", controllers.DownloadArchive)
				// 下载文件
				file.GET("download/:id", controllers.Download)
			}
		}

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
				// 创建文件下载会话
				file.PUT("download/*path", controllers.CreateDownloadSession)
				// 预览文件
				file.GET("preview/*path", controllers.Preview)
				// 获取缩略图
				file.GET("thumb/:id", controllers.Thumb)
				// 取得文件外链
				file.GET("source/:id", controllers.GetSource)
				// 打包要下载的文件
				file.POST("archive", controllers.Archive)
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
