package controllers

import (
	"context"
	ariaCall "github.com/HFO4/cloudreve/pkg/aria2"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/service/aria2"
	"github.com/HFO4/cloudreve/service/explorer"
	"github.com/gin-gonic/gin"
	"strings"
)

// AddAria2URL 添加离线下载URL
func AddAria2URL(c *gin.Context) {
	var addService aria2.AddURLService
	if err := c.ShouldBindJSON(&addService); err == nil {
		res := addService.Add(c, ariaCall.URLTask)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// SelectAria2File 选择多文件离线下载中要下载的文件
func SelectAria2File(c *gin.Context) {
	var selectService aria2.SelectFileService
	if err := c.ShouldBindJSON(&selectService); err == nil {
		res := selectService.Select(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// AddAria2Torrent 添加离线下载种子
func AddAria2Torrent(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.SingleFileService
	if err := c.ShouldBindUri(&service); err == nil {
		// 验证必须是种子文件
		filePath := c.Param("path")
		if !strings.HasSuffix(filePath, ".torrent") {
			c.JSON(200, serializer.ParamErr("只能下载 .torrent 文件", nil))
			return
		}

		// 获取种子内容的下载地址
		res := service.CreateDownloadSession(ctx, c)
		if res.Code != 0 {
			c.JSON(200, res)
			return
		}

		// 创建下载任务
		var addService aria2.AddURLService
		addService.URL = res.Data.(string)

		if err := c.ShouldBindJSON(&addService); err == nil {
			addService.URL = res.Data.(string)
			res := addService.Add(c, ariaCall.URLTask)
			c.JSON(200, res)
		} else {
			c.JSON(200, ErrorResponse(err))
		}

	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// CancelAria2Download 取消aria2离线下载任务
func CancelAria2Download(c *gin.Context) {
	var selectService aria2.DownloadTaskService
	if err := c.ShouldBindUri(&selectService); err == nil {
		res := selectService.Delete(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// ListDownloading 获取正在下载中的任务
func ListDownloading(c *gin.Context) {
	var service aria2.DownloadListService
	if err := c.ShouldBindQuery(&service); err == nil {
		res := service.Downloading(c, CurrentUser(c))
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}
