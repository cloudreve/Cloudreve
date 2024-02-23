package controllers

import (
	"context"
	"net/http"
	"sync"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
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
		Mutex:      &sync.Mutex{},
	}
}

// ServeWebDAV 处理WebDAV相关请求
func ServeWebDAV(c *gin.Context) {
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		util.Log().Warning("Failed to initialize filesystem for WebDAV，%s", err)
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

		// 检查是否只读
		if application.Readonly {
			switch c.Request.Method {
			case "DELETE", "PUT", "MKCOL", "COPY", "MOVE":
				c.Status(http.StatusForbidden)
				return
			}
		}

		// 更新Context
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), fsctx.WebDAVCtx, application))
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

// UpdateWebDAVAccounts 更改WebDAV账户只读性和是否使用代理服务
func UpdateWebDAVAccounts(c *gin.Context) {
	var service setting.WebDAVAccountUpdateService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Update(c, CurrentUser(c))
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// DeleteWebDAVMounts 删除WebDAV挂载
func DeleteWebDAVMounts(c *gin.Context) {
	var service setting.WebDAVListService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.Unmount(c, CurrentUser(c))
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// UpdateWebDAVAccountsReadonly 更改WebDAV账户只读性
func UpdateWebDAVAccountsReadonly(c *gin.Context) {
	var service setting.WebDAVAccountUpdateReadonlyService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Update(c, CurrentUser(c))
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

// CreateWebDAVMounts 创建WebDAV目录挂载
func CreateWebDAVMounts(c *gin.Context) {
	var service setting.WebDAVMountCreateService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Create(c, CurrentUser(c))
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}
