package controllers

import (
	"github.com/HFO4/cloudreve/service/share"
	"github.com/gin-gonic/gin"
)

// CreateShare 创建分享
func CreateShare(c *gin.Context) {
	var service share.ShareCreateService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Create(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// GetShare 查看分享
func GetShare(c *gin.Context) {
	var service share.ShareGetService
	if err := c.ShouldBindQuery(&service); err == nil {
		res := service.Get(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// GetShareDownload 创建分享下载会话
func GetShareDownload(c *gin.Context) {
	var service share.SingleFileService
	if err := c.ShouldBindQuery(&service); err == nil {
		res := service.CreateDownloadSession(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}
