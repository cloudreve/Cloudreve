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
		Prefix:     "/dav/",
		LockSystem: make(map[uint]webdav.LockSystem),
	}
}

func ServeWebDAV(c *gin.Context) {
	// 测试用user
	user, _ := model.GetUserByID(1)
	c.Set("user", &user)
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		util.Log().Panic("%s", err)
	}

	handler.ServeHTTP(c.Writer, c.Request, fs)
}
