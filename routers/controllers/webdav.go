package controllers

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/HFO4/cloudreve/pkg/webdav"
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
