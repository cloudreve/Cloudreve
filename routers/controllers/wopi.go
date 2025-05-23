package controllers

import (
	"net/http"

	"github.com/cloudreve/Cloudreve/v4/pkg/wopi"
	"github.com/cloudreve/Cloudreve/v4/service/explorer"
	"github.com/gin-gonic/gin"
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
	service := &explorer.WopiService{}
	err := service.PutContent(c, false)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Header(wopi.ServerErrorHeader, err.Error())
	}
}

// ModifyFile Modify file properties
func ModifyFile(c *gin.Context) {
	action := c.GetHeader(wopi.OverwriteHeader)
	var (
		service explorer.WopiService
		err     error
	)

	switch action {
	case wopi.MethodLock:
		err = service.Lock(c)
		if err == nil {
			return
		}
	case wopi.MethodRefreshLock:
		err = service.RefreshLock(c)
		if err == nil {
			return
		}
	case wopi.MethodUnlock:
		err = service.Unlock(c)
		if err == nil {
			return
		}
	case wopi.MethodPutRelative:
		err = service.PutContent(c, true)
		if err == nil {
			return
		}
	default:
		c.Status(http.StatusNotImplemented)
		return
	}

	c.Status(http.StatusInternalServerError)
	c.Header(wopi.ServerErrorHeader, err.Error())
	return
}
