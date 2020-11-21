package controllers

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/cloudreve/Cloudreve/v3/pkg/webdav"
	"github.com/cloudreve/Cloudreve/v3/service/setting"
	"github.com/gin-gonic/gin"
)

var handler *webdav.Handler

func init() {
	handler = &webdav.Handler{
		Prefix:     "/dav",
		LockSystem: make(map[uint]webdav.LockSystem),
	}
}

// ServeWebDAV 处理WebDAV相关请求
func ServeWebDAV(c *gin.Context) {
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		util.Log().Warning("无法为WebDAV初始化文件系统，%s", err)
		return
	}

	if webdavCtx, ok := c.Get("webdav"); ok {
		application := webdavCtx.(*model.Webdav)

		// 重定根目录
		if application.Root != "/" {
			if exist, root := fs.IsPathExist(application.Root); exist {
				root.Position = ""
				root.Name = "/"
				fs.Root = root
			}
		}
	}

	handler.ServeHTTP(c.Writer, c.Request, fs)
}

// GetWebDAVAccounts 获取webdav账号列表
func GetWebDAVAccounts(c *gin.Context) {
	var service setting.WebDAVListService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.Accounts(c, CurrentUser(c))
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// DeleteWebDAVAccounts 删除WebDAV账户
func DeleteWebDAVAccounts(c *gin.Context) {
	var service setting.WebDAVAccountService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.Delete(c, CurrentUser(c))
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// CreateWebDAVAccounts 创建WebDAV账户
func CreateWebDAVAccounts(c *gin.Context) {
	var service setting.WebDAVAccountCreateService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Create(c, CurrentUser(c))
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}
