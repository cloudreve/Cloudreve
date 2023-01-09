package controllers

import (
	"context"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/wopi"
	"github.com/cloudreve/Cloudreve/v3/service/explorer"
	"github.com/gin-gonic/gin"
	"net/http"
)

// CheckFileInfo Get file info
func CheckFileInfo(c *gin.Context) {
	var service explorer.WopiService
	res, err := service.FileInfo(c)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Header(wopi.ServerErrorHeader, err.Error())
		return
	}

	c.JSON(200, res)
}

// GetFile Get file content
func GetFile(c *gin.Context) {
	var service explorer.WopiService
	err := service.GetFile(c)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Header(wopi.ServerErrorHeader, err.Error())
		return
	}
}

// PutFile Puts file content
func PutFile(c *gin.Context) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	service := &explorer.FileIDService{}
	res := service.PutContent(ctx, c)
	switch res.Code {
	case serializer.CodeFileTooLarge:
		c.Status(http.StatusRequestEntityTooLarge)
		c.Header(wopi.ServerErrorHeader, res.Error)
	case serializer.CodeNotFound:
		c.Status(http.StatusNotFound)
		c.Header(wopi.ServerErrorHeader, res.Error)
	case 0:
		c.Status(http.StatusOK)
	default:
		c.Status(http.StatusInternalServerError)
		c.Header(wopi.ServerErrorHeader, res.Error)
	}
}

// ModifyFile Modify file properties
func ModifyFile(c *gin.Context) {
	action := c.GetHeader(wopi.OverwriteHeader)
	switch action {
	case wopi.MethodLock, wopi.MethodRefreshLock, wopi.MethodUnlock:
		c.Status(http.StatusOK)
		return
	case wopi.MethodRename:
		var service explorer.WopiService
		err := service.Rename(c)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			c.Header(wopi.ServerErrorHeader, err.Error())
			return
		}
	default:
		c.Status(http.StatusNotImplemented)
		return
	}
}
