package controllers

import (
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/service/callback"
	"github.com/gin-gonic/gin"
)

// RemoteCallback 远程上传回调
func RemoteCallback(c *gin.Context) {
	var callbackBody callback.RemoteUploadCallbackService
	if err := c.ShouldBindJSON(&callbackBody); err == nil {
		res := callback.ProcessCallback(callbackBody, c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// QiniuCallback 七牛上传回调
func QiniuCallback(c *gin.Context) {
	var callbackBody callback.UploadCallbackService
	if err := c.ShouldBindJSON(&callbackBody); err == nil {
		res := callback.ProcessCallback(callbackBody, c)
		if res.Code != 0 {
			c.JSON(401, serializer.QiniuCallbackFailed{Error: res.Msg})
		} else {
			c.JSON(200, res)
		}
	} else {
		c.JSON(401, ErrorResponse(err))
	}
}

// OSSCallback 阿里云OSS上传回调
func OSSCallback(c *gin.Context) {
	var callbackBody callback.UploadCallbackService
	if err := c.ShouldBindJSON(&callbackBody); err == nil {
		res := callback.ProcessCallback(callbackBody, c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}
