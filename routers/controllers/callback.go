package controllers

import (
	"github.com/HFO4/cloudreve/service/callback"
	"github.com/gin-gonic/gin"
)

// RemoteCallback 远程上传回调
func RemoteCallback(c *gin.Context) {
	var callbackBody callback.RemoteUploadCallbackService
	if err := c.ShouldBindJSON(&callbackBody); err == nil {
		res := callbackBody.Process(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}
