package controllers

import (
	"context"

	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/common"
	"github.com/cloudreve/Cloudreve/v3/service/aria2"
	"github.com/cloudreve/Cloudreve/v3/service/explorer"
	"github.com/gin-gonic/gin"
)

// AddAria2URL 添加离线下载URL
func AddAria2URL(c *gin.Context) {
	var addService aria2.AddURLService
	if err := c.ShouldBindJSON(&addService); err == nil {
		res := addService.Add(c, common.URLTask)
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

	var service explorer.FileIDService
	if err := c.ShouldBindUri(&service); err == nil {
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
			res := addService.Add(c, common.URLTask)
			c.JSON(200, res)
		} else {
			c.JSON(200, ErrorResponse(err))
		}

	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// CancelAria2Download 取消或删除aria2离线下载任务
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

// ListFinished 获取已完成的任务
func ListFinished(c *gin.Context) {
	var service aria2.DownloadListService
	if err := c.ShouldBindQuery(&service); err == nil {
		res := service.Finished(c, CurrentUser(c))
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}
