package routers

import (
	"github.com/HFO4/cloudreve/middleware"
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/HFO4/cloudreve/routers/controllers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// InitRouter 初始化路由
func InitRouter() *gin.Engine {
	if conf.SystemConfig.Mode == "master" {
		util.Log().Info("当前运行模式：Master")
		return InitMasterRouter()
	}
	util.Log().Info("当前运行模式：Slave")
	return InitSlaveRouter()

}

// InitSlaveRouter 初始化从机模式路由
func InitSlaveRouter() *gin.Engine {
	r := gin.Default()
	// 跨域相关
	InitCORS(r)
	v3 := r.Group("/api/v3/slave")
	// 鉴权中间件
	v3.Use(middleware.SignRequired())

	/*
		路由
	*/
	{
		// 上传
		v3.POST("upload", controllers.SlaveUpload)
		// 下载
		v3.GET("download/:speed/:path/:name", controllers.SlaveDownload)
		// 预览 / 外链
		v3.GET("source/:speed/:path/:name", controllers.SlavePreview)
		// 缩略图
		v3.GET("thumb/:path", controllers.SlaveThumb)
		// 删除文件
		v3.POST("delete", controllers.SlaveDelete)
	}
	return r
}

// InitCORS 初始化跨域配置
func InitCORS(router *gin.Engine) {
	if conf.CORSConfig.AllowOrigins[0] != "UNSET" {
		router.Use(cors.New(cors.Config{
			AllowOrigins:     conf.CORSConfig.AllowOrigins,
			AllowMethods:     conf.CORSConfig.AllowMethods,
			AllowHeaders:     conf.CORSConfig.AllowHeaders,
			AllowCredentials: conf.CORSConfig.AllowCredentials,
			ExposeHeaders:    conf.CORSConfig.ExposeHeaders,
		}))
		return
	}

	// slave模式下未启动跨域的警告
	if conf.SystemConfig.Mode == "slave" {
		util.Log().Warning("当前作为存储端（Slave）运行，但未启用跨域配置，可能会导致 Master 端无法正常上传文件")
	}
}

// InitMasterRouter 初始化主机模式路由
func InitMasterRouter() *gin.Engine {
	r := gin.Default()
	v3 := r.Group("/api/v3")
	/*
		中间件
	*/
	v3.Use(middleware.Session(conf.SystemConfig.SessionSecret))
	// 跨域相关
	InitCORS(r)
	// 测试模式加入Mock助手中间件
	if gin.Mode() == gin.TestMode {
		v3.Use(middleware.MockHelper())
	}
	v3.Use(middleware.CurrentUser())

	/*
		路由
	*/
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

		// 回调接口
		callback := v3.Group("callback")
		{
			// 远程策略上传回调
			callback.POST(
				"remote/:key",
				middleware.RemoteCallbackAuth(),
				controllers.RemoteCallback,
			)
			// 七牛策略上传回调
			callback.POST(
				"qiniu/:key",
				middleware.QiniuCallbackAuth(),
				controllers.QiniuCallback,
			)
			// 阿里云OSS策略上传回调
			callback.POST(
				"oss/:key",
				middleware.OSSCallbackAuth(),
				controllers.OSSCallback,
			)
			// 又拍云策略上传回调
			callback.POST(
				"upyun/:key",
				middleware.UpyunCallbackAuth(),
				controllers.UpyunCallback,
			)
			onedrive := callback.Group("onedrive")
			{
				// 文件上传完成
				onedrive.POST(
					"finish/:key",
					middleware.OneDriveCallbackAuth(),
					controllers.OneDriveCallback,
				)
			}
			// 腾讯云COS策略上传回调
			callback.GET(
				"cos/:key",
				middleware.COSCallbackAuth(),
				controllers.COSCallback,
			)
		}

		// 分享相关
		share := v3.Group("share", middleware.ShareAvailable())
		{
			// 获取分享
			share.GET("info/:id", controllers.GetShare)
			// 创建文件下载会话
			share.PUT("download/:id",
				middleware.CheckShareUnlocked(),
				middleware.BeforeShareDownload(),
				controllers.GetShareDownload,
			)
			// 预览分享文件
			share.GET("preview/:id",
				middleware.CheckShareUnlocked(),
				middleware.ShareCanPreview(),
				middleware.BeforeShareDownload(),
				controllers.PreviewShare,
			)
			// 取得Office文档预览地址
			share.GET("doc/:id",
				middleware.CheckShareUnlocked(),
				middleware.ShareCanPreview(),
				middleware.BeforeShareDownload(),
				controllers.GetShareDocPreview,
			)
			// 获取文本文件内容
			share.GET("content/:id",
				middleware.CheckShareUnlocked(),
				middleware.BeforeShareDownload(),
				controllers.PreviewShareText,
			)
			// 分享目录列文件
			share.GET("list/:id/*path",
				middleware.CheckShareUnlocked(),
				controllers.ListSharedFolder,
			)
			// 归档打包下载
			share.POST("archive/:id",
				middleware.CheckShareUnlocked(),
				middleware.BeforeShareDownload(),
				controllers.ArchiveShare,
			)
			// 获取README文本文件内容
			share.GET("readme/:id",
				middleware.CheckShareUnlocked(),
				controllers.PreviewShareReadme,
			)
			// 获取缩略图
			share.GET("thumb/:id/:file",
				middleware.CheckShareUnlocked(),
				middleware.ShareCanPreview(),
				controllers.ShareThumb,
			)
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
				// 存储信息
				user.GET("storage", controllers.UserStorage)
				// 退出登录
				user.DELETE("session", controllers.UserSignOut)

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
				// 获取上传凭证
				file.GET("upload/credential", controllers.GetUploadCredential)
				// 更新文件
				file.PUT("update/*path", controllers.PutContent)
				// 创建文件下载会话
				file.PUT("download/*path", controllers.CreateDownloadSession)
				// 预览文件
				file.GET("preview/*path", controllers.Preview)
				// 获取文本文件内容
				file.GET("content/*path", controllers.PreviewText)
				// 取得Office文档预览地址
				file.GET("doc/*path", controllers.GetDocPreview)
				// 获取缩略图
				file.GET("thumb/:id", controllers.Thumb)
				// 取得文件外链
				file.GET("source/:id", controllers.GetSource)
				// 打包要下载的文件
				file.POST("archive", controllers.Archive)
				// 创建文件压缩任务
				file.POST("compress", controllers.Compress)
				// 创建文件解压缩任务
				file.POST("decompress", controllers.Decompress)
			}

			// 离线下载任务
			aria2 := auth.Group("aria2")
			{
				// 创建URL下载任务
				aria2.POST("url", controllers.AddAria2URL)
				// 创建种子下载任务
				aria2.POST("torrent/*path", controllers.AddAria2Torrent)
				// 重新选择要下载的文件
				aria2.PUT("select/:gid", controllers.SelectAria2File)
				// 取消下载任务
				aria2.DELETE("task/:gid", controllers.CancelAria2Download)
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

			// 分享
			share := auth.Group("share")
			{
				// 创建新分享
				share.POST("", controllers.CreateShare)
				// 转存他人分享
				share.POST("save/:id",
					middleware.ShareAvailable(),
					middleware.CheckShareUnlocked(),
					middleware.BeforeShareDownload(),
					controllers.SaveShare,
				)
			}

		}

	}

	// 初始化WebDAV相关路由
	initWebDAV(r.Group("dav"))
	return r
}

// initWebDAV 初始化WebDAV相关路由
func initWebDAV(group *gin.RouterGroup) {
	{
		group.Use(middleware.WebDAVAuth())

		group.Any("/*path", controllers.ServeWebDAV)
		group.Any("", controllers.ServeWebDAV)
		group.Handle("PROPFIND", "/*path", controllers.ServeWebDAV)
		group.Handle("PROPFIND", "", controllers.ServeWebDAV)
		group.Handle("MKCOL", "/*path", controllers.ServeWebDAV)
		group.Handle("LOCK", "/*path", controllers.ServeWebDAV)
		group.Handle("UNLOCK", "/*path", controllers.ServeWebDAV)
		group.Handle("PROPPATCH", "/*path", controllers.ServeWebDAV)
		group.Handle("COPY", "/*path", controllers.ServeWebDAV)
		group.Handle("MOVE", "/*path", controllers.ServeWebDAV)

	}
}
