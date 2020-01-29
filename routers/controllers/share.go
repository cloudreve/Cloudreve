package controllers

import (
	"context"
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

// PreviewShare 预览分享文件内容
func PreviewShare(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service share.SingleFileService
	if err := c.ShouldBindQuery(&service); err == nil {
		res := service.PreviewContent(ctx, c, false)
		// 是否需要重定向
		if res.Code == -301 {
			c.Redirect(301, res.Data.(string))
			return
		}
		// 是否有错误发生
		if res.Code != 0 {
			c.JSON(200, res)
		}
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// PreviewShareText 预览文本文件
func PreviewShareText(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service share.SingleFileService
	if err := c.ShouldBindQuery(&service); err == nil {
		res := service.PreviewContent(ctx, c, true)
		// 是否有错误发生
		if res.Code != 0 {
			c.JSON(200, res)
		}
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// GetShareDocPreview 创建分享Office文档预览地址
func GetShareDocPreview(c *gin.Context) {
	var service share.SingleFileService
	if err := c.ShouldBindQuery(&service); err == nil {
		res := service.CreateDocPreviewSession(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// SaveShare 转存他人分享
func SaveShare(c *gin.Context) {
	var service share.SingleFileService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.SaveToMyFile(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}
