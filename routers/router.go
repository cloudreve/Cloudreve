package routers

import (
	// "github.com/abslant/gzip"
	"github.com/cloudreve/Cloudreve/v3/bootstrap"
	"github.com/cloudreve/Cloudreve/v3/middleware"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/cluster"
	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	wopi2 "github.com/cloudreve/Cloudreve/v3/pkg/wopi"
	"github.com/cloudreve/Cloudreve/v3/routers/controllers"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// InitRouter 初始化路由
func InitRouter() *gin.Engine {
	if conf.SystemConfig.Mode == "master" {
		util.Log().Info("Current running mode: Master.")
		return InitMasterRouter()
	}
	util.Log().Info("Current running mode: Slave.")
	return InitSlaveRouter()

}

// InitSlaveRouter 初始化从机模式路由
func InitSlaveRouter() *gin.Engine {
	r := gin.Default()
	// 跨域相关
	InitCORS(r)
	v3 := r.Group("/api/v3/slave")
	// 鉴权中间件
	v3.Use(middleware.SignRequired(auth.General))
	// 主机信息解析
	v3.Use(middleware.MasterMetadata())
	// 禁止缓存
	v3.Use(middleware.CacheControl())

	/*
		路由
	*/
	{
		// Ping
		v3.POST("ping", controllers.SlavePing)
		// 测试 Aria2 RPC 连接
		v3.POST("ping/aria2", controllers.AdminTestAria2)
		// 接收主机心跳包
		v3.POST("heartbeat", controllers.SlaveHeartbeat)
		// 上传
		upload := v3.Group("upload")
		{
			// 上传分片
			upload.POST(":sessionId", controllers.SlaveUpload)
			// 创建上传会话上传
			upload.PUT("", controllers.SlaveGetUploadSession)
			// 删除上传会话
			upload.DELETE(":sessionId", controllers.SlaveDeleteUploadSession)
		}
		// 下载
		v3.GET("download/:speed/:path/:name", controllers.SlaveDownload)
		// 预览 / 外链
		v3.GET("source/:speed/:path/:name", controllers.SlavePreview)
		// 缩略图
		v3.GET("thumb/:path/:ext", controllers.SlaveThumb)
		// 删除文件
		v3.POST("delete", controllers.SlaveDelete)
		// 列出文件
		v3.POST("list", controllers.SlaveList)

		// 离线下载
		aria2 := v3.Group("aria2")
		aria2.Use(middleware.UseSlaveAria2Instance(cluster.DefaultController))
		{
			// 创建离线下载任务
			aria2.POST("task", controllers.SlaveAria2Create)
			// 获取任务状态
			aria2.POST("status", controllers.SlaveAria2Status)
			// 取消离线下载任务
			aria2.POST("cancel", controllers.SlaveCancelAria2Task)
			// 选取任务文件
			aria2.POST("select", controllers.SlaveSelectTask)
			// 删除任务临时文件
			aria2.POST("delete", controllers.SlaveDeleteTempFile)
		}

		// 异步任务
		task := v3.Group("task")
		{
			task.PUT("transfer", controllers.SlaveCreateTransferTask)
		}
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
		util.Log().Warning("You are running Cloudreve as slave node, if you are using slave storage policy, please enable CORS feature in config file, otherwise file cannot be uploaded from Master site.")
	}
}

// InitMasterRouter 初始化主机模式路由
func InitMasterRouter() *gin.Engine {
	r := gin.Default()
	bootstrap.InitCustomRoute(r.Group("/api/v3"))
	// bootstrap.InitCustomRoute(r.Group("/custom"))

	/*
		静态资源
	*/
	r.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPaths([]string{"/api/"})))
	// r.Use(gzip.GzipHandler())
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
	// 用户会话
	v3.Use(middleware.CurrentUser())

	// 禁止缓存
	v3.Use(middleware.CacheControl())

	/*
		路由
	*/
	{
		// Redirect file source link
		source := r.Group("f")
		{
			source.GET(":id/:name",
				middleware.HashID(hashid.SourceLinkID),
				middleware.ValidateSourceLink(),
				controllers.AnonymousPermLink)
		}

		// 全局设置相关
		site := v3.Group("site")
		{
			// 测试用路由
			site.GET("ping", controllers.Ping)
			// 验证码
			site.GET("captcha", controllers.Captcha)
			// 站点全局配置
			site.GET("config", middleware.CSRFInit(), controllers.SiteConfig)
			// VOL 密钥
			site.GET("vol", controllers.GetVolSecret)
		}

		// 用户相关路由
		user := v3.Group("user")
		{
			// 用户登录
			user.POST("session", middleware.CaptchaRequired("login_captcha"), controllers.UserLogin)
			// 用户注册
			user.POST("",
				middleware.IsFunctionEnabled("register_enabled"),
				middleware.CaptchaRequired("reg_captcha"),
				controllers.UserRegister,
			)
			// 用二步验证户登录
			user.POST("2fa", controllers.User2FALogin)
			// 发送密码重设邮件
			user.POST("reset", middleware.CaptchaRequired("forget_captcha"), controllers.UserSendReset)
			// 通过邮件里的链接重设密码
			user.PATCH("reset", controllers.UserReset)
			// 邮件激活
			user.GET("activate/:id",
				middleware.SignRequired(auth.General),
				middleware.HashID(hashid.UserID),
				controllers.UserActivate,
			)
			// 初始化QQ登录
			user.POST("qq", controllers.UserQQLogin)
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
				middleware.StaticResourceCache(),
				controllers.GetUserAvatar,
			)
		}

		// 需要携带签名验证的
		sign := v3.Group("")
		sign.Use(middleware.SignRequired(auth.General))
		{
			file := sign.Group("file")
			{
				// 文件外链（直接输出文件数据）
				file.GET("get/:id/:name",
					middleware.Sandbox(),
					middleware.StaticResourceCache(),
					controllers.AnonymousGetContent,
				)
				// 文件外链(301跳转)
				file.GET("source/:id/:name", controllers.AnonymousPermLinkDeprecated)
				// 下载文件
				file.GET("download/:id",
					middleware.StaticResourceCache(),
					controllers.Download,
				)
				// 打包并下载文件
				file.GET("archive/:sessionID/archive.zip", controllers.DownloadArchive)
			}

			// Copy user session
			sign.GET(
				"user/session/copy/:id",
				middleware.MobileRequestOnly(),
				controllers.UserPerformCopySession,
			)
		}

		// 从机的 RPC 通信
		slave := v3.Group("slave")
		slave.Use(middleware.SlaveRPCSignRequired(cluster.Default))
		{
			// 事件通知
			slave.PUT("notification/:subject", controllers.SlaveNotificationPush)
			// 上传
			upload := slave.Group("upload")
			{
				// 上传分片
				upload.POST(":sessionId", controllers.SlaveUpload)
				// 创建上传会话上传
				upload.PUT("", controllers.SlaveGetUploadSession)
				// 删除上传会话
				upload.DELETE(":sessionId", controllers.SlaveDeleteUploadSession)
			}
			// Oauth 存储策略凭证
			slave.GET("credential/:id", controllers.SlaveGetOauthCredential)
		}

		// 回调接口
		callback := v3.Group("callback")
		{
			// QQ互联回调
			callback.POST(
				"qq",
				controllers.QQCallback,
			)
			// PAYJS回调
			callback.POST(
				"payjs",
				controllers.PayJSCallback,
			)
			// 支付宝回调
			callback.POST(
				"alipay",
				controllers.AlipayCallback,
			)
			// 微信扫码支付回调
			callback.POST(
				"wechat",
				controllers.WechatCallback,
			)
			// Custom payment callback
			callback.GET(
				"custom/:orderno/:id",
				middleware.SignRequired(auth.General),
				controllers.CustomCallback,
			)
			// 远程策略上传回调
			callback.POST(
				"remote/:sessionID/:key",
				middleware.UseUploadSession("remote"),
				middleware.RemoteCallbackAuth(),
				controllers.RemoteCallback,
			)
			// 七牛策略上传回调
			callback.POST(
				"qiniu/:sessionID",
				middleware.UseUploadSession("qiniu"),
				middleware.QiniuCallbackAuth(),
				controllers.QiniuCallback,
			)
			// 阿里云OSS策略上传回调
			callback.POST(
				"oss/:sessionID",
				middleware.UseUploadSession("oss"),
				middleware.OSSCallbackAuth(),
				controllers.OSSCallback,
			)
			// 又拍云策略上传回调
			callback.POST(
				"upyun/:sessionID",
				middleware.UseUploadSession("upyun"),
				middleware.UpyunCallbackAuth(),
				controllers.UpyunCallback,
			)
			onedrive := callback.Group("onedrive")
			{
				// 文件上传完成
				onedrive.POST(
					"finish/:sessionID",
					middleware.UseUploadSession("onedrive"),
					middleware.OneDriveCallbackAuth(),
					controllers.OneDriveCallback,
				)
				// OAuth 完成
				onedrive.GET(
					"auth",
					controllers.OneDriveOAuth,
				)
			}
			// Google Drive related
			gdrive := callback.Group("googledrive")
			{
				// OAuth 完成
				gdrive.GET(
					"auth",
					controllers.GoogleDriveOAuth,
				)
			}
			// 腾讯云COS策略上传回调
			callback.GET(
				"cos/:sessionID",
				middleware.UseUploadSession("cos"),
				controllers.COSCallback,
			)
			// AWS S3策略上传回调
			callback.GET(
				"s3/:sessionID",
				middleware.UseUploadSession("s3"),
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
			// 分享目录搜索
			share.GET("search/:id/:type/:keywords",
				middleware.CheckShareUnlocked(),
				controllers.SearchSharedFolder,
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
			// 举报分享
			share.POST("report/:id",
				middleware.IsFunctionEnabled("report_enabled"),
				middleware.CheckShareUnlocked(),
				controllers.ReportShare,
			)
			// 搜索公共分享
			v3.Group("share").GET("search", middleware.AuthRequired(), controllers.SearchShare)
		}

		wopi := v3.Group(
			"wopi",
			middleware.HashID(hashid.FileID),
			middleware.WopiAccessValidation(wopi2.Default, cache.Store),
		)
		{
			// 获取文件信息
			wopi.GET("files/:id", controllers.CheckFileInfo)
			// 获取文件内容
			wopi.GET("files/:id/contents", controllers.GetFile)
			// 更新文件内容
			wopi.POST("files/:id/contents", middleware.WopiWriteAccess(), controllers.PutFile)
			// 通用文件操作
			wopi.POST("files/:id", middleware.WopiWriteAccess(), controllers.ModifyFile)
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
				// admin.GET("news", controllers.AdminNews)
				// 更改设置
				admin.PATCH("setting", controllers.AdminChangeSetting)
				// 获取设置
				admin.POST("setting", controllers.AdminGetSetting)
				// 获取用户组列表
				admin.GET("groups", controllers.AdminGetGroups)
				// 重新加载子服务
				admin.GET("reload/:service", controllers.AdminReloadService)
				// 测试设置
				test := admin.Group("test")
				{
					// 测试邮件设置
					test.POST("mail", controllers.AdminSendTestMail)
					// 测试缩略图生成器调用
					test.POST("thumb", controllers.AdminTestThumbGenerator)
				}

				// 离线下载相关
				aria2 := admin.Group("aria2")
				{
					// 测试连接配置
					aria2.POST("test", controllers.AdminTestAria2)
				}

				vol := admin.Group("vol")
				{
					vol.GET("sync", controllers.AdminSyncVol)
				}

				// 兑换码相关
				redeem := admin.Group("redeem")
				{
					// 列出激活码
					redeem.POST("list", controllers.AdminListRedeems)
					// 生成激活码
					redeem.POST("", controllers.AdminGenerateRedeems)
					// 删除激活码
					redeem.DELETE(":id", controllers.AdminDeleteRedeem)
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
					oauth := policy.Group(":id/oauth")
					{
						// 获取 OneDrive OAuth URL
						oauth.GET("onedrive", controllers.AdminOAuthURL("onedrive"))
						// 获取 Google Drive OAuth URL
						oauth.GET("googledrive", controllers.AdminOAuthURL("googledrive"))
					}

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
					file.GET("preview/:id", middleware.Sandbox(), controllers.AdminGetFile)
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

				order := admin.Group("order")
				{
					// 列出订单
					order.POST("list", controllers.AdminListOrder)
					// 删除
					order.POST("delete", controllers.AdminDeleteOrder)
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

				report := admin.Group("report")
				{
					// 列出未处理举报
					report.POST("list", controllers.AdminListReport)
					// 删除
					report.POST("delete", controllers.AdminDeleteReport)
				}

				node := admin.Group("node")
				{
					// 列出从机节点
					node.POST("list", controllers.AdminListNodes)
					// 列出从机节点
					node.POST("aria2/test", controllers.AdminTestAria2)
					// 创建/保存节点
					node.POST("", controllers.AdminAddNode)
					// 启用/暂停节点
					node.PATCH("enable/:id/:desired", controllers.AdminToggleNode)
					// 删除节点
					node.DELETE(":id", controllers.AdminDeleteNode)
					// 获取节点
					node.GET(":id", controllers.AdminGetNode)
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
				// Generate temp URL for copying client-side session, used in adding accounts
				// for mobile App.
				user.GET("session", controllers.UserPrepareCopySession)

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
					// 获取用户可选存储策略
					setting.GET("policies", controllers.UserAvailablePolicies)
					// 获取用户可选节点
					setting.GET("nodes", controllers.UserAvailableNodes)
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
			file := auth.Group("file", middleware.PhoneRequired(), middleware.HashID(hashid.FileID))
			{
				// 上传
				upload := file.Group("upload")
				{
					// 文件上传
					upload.POST(":sessionId/:index", controllers.FileUpload)
					// 创建上传会话
					upload.PUT("", controllers.GetUploadSession)
					// 删除给定上传会话
					upload.DELETE(":sessionId", controllers.DeleteUploadSession)
					// 删除全部上传会话
					upload.DELETE("", controllers.DeleteAllUploadSession)
				}
				// 更新文件
				file.PUT("update/:id", controllers.PutContent)
				// 创建空白文件
				file.POST("create", controllers.CreateFile)
				// 创建文件下载会话
				file.PUT("download/:id", controllers.CreateDownloadSession)
				// 预览文件
				file.GET("preview/:id", middleware.Sandbox(), controllers.Preview)
				// 获取文本文件内容
				file.GET("content/:id", middleware.Sandbox(), controllers.PreviewText)
				// 取得Office文档预览地址
				file.GET("doc/:id", controllers.GetDocPreview)
				// 获取缩略图
				file.GET("thumb/:id", controllers.Thumb)
				// 取得文件外链
				file.POST("source", controllers.GetSource)
				// 打包要下载的文件
				file.POST("archive", controllers.Archive)
				// 创建文件压缩任务
				file.POST("compress", controllers.Compress)
				// 创建文件解压缩任务
				file.POST("decompress", controllers.Decompress)
				// 创建文件转移任务
				file.POST("relocate", controllers.Relocate)
				// 搜索文件
				file.GET("search/:type/:keywords", controllers.SearchFile)
			}

			// 离线下载任务
			aria2 := auth.Group("aria2", middleware.PhoneRequired())
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
			directory := auth.Group("directory", middleware.PhoneRequired())
			{
				// 创建目录
				directory.PUT("", controllers.CreateDirectory)
				// 列出目录下内容
				directory.GET("*path", controllers.ListDirectory)
			}

			// 对象，文件和目录的抽象
			object := auth.Group("object", middleware.PhoneRequired())
			{
				// 删除对象
				object.DELETE("", controllers.Delete)
				// 移动对象
				object.PATCH("", controllers.Move)
				// 复制对象
				object.POST("copy", controllers.Copy)
				// 重命名对象
				object.POST("rename", controllers.Rename)
				// 获取对象属性
				object.GET("property/:id", controllers.GetProperty)
			}

			// 分享
			share := auth.Group("share", middleware.PhoneRequired())
			{
				// 创建新分享
				share.POST("", controllers.CreateShare)
				// 列出我的分享
				share.GET("", controllers.ListShare)
				// 转存他人分享
				share.POST("save/:id",
					middleware.ShareAvailable(),
					middleware.CheckShareUnlocked(),
					middleware.BeforeShareDownload(),
					controllers.SaveShare,
				)
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

			// 增值服务相关
			vas := auth.Group("vas", middleware.PhoneRequired())
			{
				// 获取容量包及配额信息
				vas.GET("pack", controllers.GetQuota)
				// 获取商品信息，同时返回支付信息
				vas.GET("product", controllers.GetProduct)
				// 新建支付订单
				vas.POST("order", controllers.NewOrder)
				// 查询订单状态
				vas.GET("order/:id", controllers.OrderStatus)
				// 获取兑换码信息
				vas.GET("redeem/:code", controllers.GetRedeemInfo)
				// 执行兑换
				vas.POST("redeem/:code", controllers.DoRedeem)
			}

			// WebDAV管理相关
			webdav := auth.Group("webdav", middleware.PhoneRequired())
			{
				// 获取账号信息
				webdav.GET("accounts", controllers.GetWebDAVAccounts)
				// 新建账号
				webdav.POST("accounts", controllers.CreateWebDAVAccounts)
				// 删除账号
				webdav.DELETE("accounts/:id", controllers.DeleteWebDAVAccounts)
				// 删除目录挂载
				webdav.DELETE("mount/:id",
					middleware.HashID(hashid.FolderID),
					controllers.DeleteWebDAVMounts,
				)
				// 创建目录挂载
				webdav.POST("mount", controllers.CreateWebDAVMounts)
				// 更新账号可读性和是否使用代理服务
				webdav.PATCH("accounts", controllers.UpdateWebDAVAccounts)
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
