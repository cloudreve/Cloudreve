package routers

import (
	"github.com/cloudreve/Cloudreve/v3/middleware"
	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/cloudreve/Cloudreve/v3/routers/controllers"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
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
		// Ping
		v3.POST("ping", controllers.SlavePing)
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
		// 列出文件
		v3.POST("list", controllers.SlaveList)
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

	/*
		静态资源
	*/
	r.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPaths([]string{"/api/"})))
	r.Use(middleware.FrontendFileHandler())
	r.GET("manifest.json", controllers.Manifest)

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
			site.GET("config", middleware.CSRFInit(), controllers.SiteConfig)
		}

		// 用户相关路由
		user := v3.Group("user")
		{
			// 用户登录
			user.POST("session", controllers.UserLogin)
			// 用户注册
			user.POST("",
				middleware.IsFunctionEnabled("register_enabled"),
				controllers.UserRegister,
			)
			// 用二步验证户登录
			user.POST("2fa", controllers.User2FALogin)
			// 发送密码重设邮件
			user.POST("reset", controllers.UserSendReset)
			// 通过邮件里的链接重设密码
			user.PATCH("reset", controllers.UserReset)
			// 邮件激活
			user.GET("activate/:id",
				middleware.SignRequired(),
				middleware.HashID(hashid.UserID),
				controllers.UserActivate,
			)
			// WebAuthn登陆初始化
			user.GET("authn/:username",
				middleware.IsFunctionEnabled("authn_enabled"),
				controllers.StartLoginAuthn,
			)
			// WebAuthn登陆
			user.POST("authn/finish/:username",
				middleware.IsFunctionEnabled("authn_enabled"),
				controllers.FinishLoginAuthn,
			)
			// 获取用户主页展示用分享
			user.GET("profile/:id",
				middleware.HashID(hashid.UserID),
				controllers.GetUserShare,
			)
			// 获取用户头像
			user.GET("avatar/:id/:size",
				middleware.HashID(hashid.UserID),
				controllers.GetUserAvatar,
			)
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
				// 文件上传完成
				onedrive.GET(
					"auth",
					controllers.OneDriveOAuth,
				)
			}
			// 腾讯云COS策略上传回调
			callback.GET(
				"cos/:key",
				middleware.COSCallbackAuth(),
				controllers.COSCallback,
			)
			// AWS S3策略上传回调
			callback.GET(
				"s3/:key",
				middleware.S3CallbackAuth(),
				controllers.S3Callback,
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
				middleware.CSRFCheck(),
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
			// 搜索公共分享
			v3.Group("share").GET("search", controllers.SearchShare)
		}

		// 需要登录保护的
		auth := v3.Group("")
		auth.Use(middleware.AuthRequired())
		{
			// 管理
			admin := auth.Group("admin", middleware.IsAdmin())
			{
				// 获取站点概况
				admin.GET("summary", controllers.AdminSummary)
				// 获取社区新闻
				admin.GET("news", controllers.AdminNews)
				// 更改设置
				admin.PATCH("setting", controllers.AdminChangeSetting)
				// 获取设置
				admin.POST("setting", controllers.AdminGetSetting)
				// 获取用户组列表
				admin.GET("groups", controllers.AdminGetGroups)
				// 重新加载子服务
				admin.GET("reload/:service", controllers.AdminReloadService)
				// 重新加载子服务
				admin.POST("mailTest", controllers.AdminSendTestMail)

				// 离线下载相关
				aria2 := admin.Group("aria2")
				{
					// 测试连接配置
					aria2.POST("test", controllers.AdminTestAria2)
				}

				// 存储策略管理
				policy := admin.Group("policy")
				{
					// 列出存储策略
					policy.POST("list", controllers.AdminListPolicy)
					// 测试本地路径可用性
					policy.POST("test/path", controllers.AdminTestPath)
					// 测试从机通信
					policy.POST("test/slave", controllers.AdminTestSlave)
					// 创建存储策略
					policy.POST("", controllers.AdminAddPolicy)
					// 创建跨域策略
					policy.POST("cors", controllers.AdminAddCORS)
					// 创建COS回调函数
					policy.POST("scf", controllers.AdminAddSCF)
					// 获取 OneDrive OAuth URL
					policy.GET(":id/oauth", controllers.AdminOneDriveOAuth)
					// 获取 存储策略
					policy.GET(":id", controllers.AdminGetPolicy)
					// 删除 存储策略
					policy.DELETE(":id", controllers.AdminDeletePolicy)
				}

				// 用户组管理
				group := admin.Group("group")
				{
					// 列出用户组
					group.POST("list", controllers.AdminListGroup)
					// 获取用户组
					group.GET(":id", controllers.AdminGetGroup)
					// 创建/保存用户组
					group.POST("", controllers.AdminAddGroup)
					// 删除
					group.DELETE(":id", controllers.AdminDeleteGroup)
				}

				user := admin.Group("user")
				{
					// 列出用户
					user.POST("list", controllers.AdminListUser)
					// 获取用户
					user.GET(":id", controllers.AdminGetUser)
					// 创建/保存用户
					user.POST("", controllers.AdminAddUser)
					// 删除
					user.POST("delete", controllers.AdminDeleteUser)
					// 封禁/解封用户
					user.PATCH("ban/:id", controllers.AdminBanUser)
				}

				file := admin.Group("file")
				{
					// 列出文件
					file.POST("list", controllers.AdminListFile)
					// 预览文件
					file.GET("preview/:id", controllers.AdminGetFile)
					// 删除
					file.POST("delete", controllers.AdminDeleteFile)
					// 列出用户或外部文件系统目录
					file.GET("folders/:type/:id/*path",
						controllers.AdminListFolders)
				}

				share := admin.Group("share")
				{
					// 列出分享
					share.POST("list", controllers.AdminListShare)
					// 删除
					share.POST("delete", controllers.AdminDeleteShare)
				}

				download := admin.Group("download")
				{
					// 列出任务
					download.POST("list", controllers.AdminListDownload)
					// 删除
					download.POST("delete", controllers.AdminDeleteDownload)
				}

				task := admin.Group("task")
				{
					// 列出任务
					task.POST("list", controllers.AdminListTask)
					// 删除
					task.POST("delete", controllers.AdminDeleteTask)
					// 新建文件导入任务
					task.POST("import", controllers.AdminCreateImportTask)
				}

			}

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
				authn := user.Group("authn",
					middleware.IsFunctionEnabled("authn_enabled"))
				{
					authn.PUT("", controllers.StartRegAuthn)
					authn.PUT("finish", controllers.FinishRegAuthn)
				}

				// 用户设置
				setting := user.Group("setting")
				{
					// 任务队列
					setting.GET("tasks", controllers.UserTasks)
					// 获取当前用户设定
					setting.GET("", controllers.UserSetting)
					// 从文件上传头像
					setting.POST("avatar", controllers.UploadAvatar)
					// 设定为Gravatar头像
					setting.PUT("avatar", controllers.UseGravatar)
					// 更改用户设定
					setting.PATCH(":option", controllers.UpdateOption)
					// 获得二步验证初始化信息
					setting.GET("2fa", controllers.UserInit2FA)
				}
			}

			// 文件
			file := auth.Group("file", middleware.HashID(hashid.FileID))
			{
				// 文件上传
				file.POST("upload", controllers.FileUploadStream)
				// 获取上传凭证
				file.GET("upload/credential", controllers.GetUploadCredential)
				// 更新文件
				file.PUT("update/:id", controllers.PutContent)
				// 创建空白文件
				file.POST("create", controllers.CreateFile)
				// 创建文件下载会话
				file.PUT("download/:id", controllers.CreateDownloadSession)
				// 预览文件
				file.GET("preview/:id", controllers.Preview)
				// 获取文本文件内容
				file.GET("content/:id", controllers.PreviewText)
				// 取得Office文档预览地址
				file.GET("doc/:id", controllers.GetDocPreview)
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
				// 创建文件解压缩任务
				file.GET("search/:type/:keywords", controllers.SearchFile)
			}

			// 离线下载任务
			aria2 := auth.Group("aria2")
			{
				// 创建URL下载任务
				aria2.POST("url", controllers.AddAria2URL)
				// 创建种子下载任务
				aria2.POST("torrent/:id", middleware.HashID(hashid.FileID), controllers.AddAria2Torrent)
				// 重新选择要下载的文件
				aria2.PUT("select/:gid", controllers.SelectAria2File)
				// 取消或删除下载任务
				aria2.DELETE("task/:gid", controllers.CancelAria2Download)
				// 获取正在下载中的任务
				aria2.GET("downloading", controllers.ListDownloading)
				// 获取已完成的任务
				aria2.GET("finished", controllers.ListFinished)
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
				// 列出我的分享
				share.GET("", controllers.ListShare)
				// 更新分享属性
				share.PATCH(":id",
					middleware.ShareAvailable(),
					middleware.ShareOwner(),
					controllers.UpdateShare,
				)
				// 删除分享
				share.DELETE(":id",
					controllers.DeleteShare,
				)
			}

			// 用户标签
			tag := auth.Group("tag")
			{
				// 创建文件分类标签
				tag.POST("filter", controllers.CreateFilterTag)
				// 创建目录快捷方式标签
				tag.POST("link", controllers.CreateLinkTag)
				// 删除标签
				tag.DELETE(":id", middleware.HashID(hashid.TagID), controllers.DeleteTag)
			}

			// WebDAV管理相关
			webdav := auth.Group("webdav")
			{
				// 获取账号信息
				webdav.GET("accounts", controllers.GetWebDAVAccounts)
				// 新建账号
				webdav.POST("accounts", controllers.CreateWebDAVAccounts)
				// 删除账号
				webdav.DELETE("accounts/:id", controllers.DeleteWebDAVAccounts)
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
