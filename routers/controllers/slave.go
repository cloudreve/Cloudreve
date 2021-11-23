package controllers

import (
	"context"
	"net/url"
	"strconv"

	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/local"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/service/admin"
	"github.com/cloudreve/Cloudreve/v3/service/aria2"
	"github.com/cloudreve/Cloudreve/v3/service/explorer"
	"github.com/cloudreve/Cloudreve/v3/service/node"
	"github.com/gin-gonic/gin"
)

// SlaveUpload 从机文件上传
func SlaveUpload(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, fsctx.GinCtx, c)
	defer cancel()

	// 创建匿名文件系统
	fs, err := filesystem.NewAnonymousFileSystem()
	if err != nil {
		c.JSON(200, serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err))
		return
	}
	fs.Handler = local.Driver{}

	// 从请求中取得上传策略
	uploadPolicyRaw := c.GetHeader("X-Cr-Policy")
	if uploadPolicyRaw == "" {
		c.JSON(200, serializer.ParamErr("未指定上传策略", nil))
		return
	}

	// 解析上传策略
	uploadPolicy, err := serializer.DecodeUploadPolicy(uploadPolicyRaw)
	if err != nil {
		c.JSON(200, serializer.ParamErr("上传策略格式有误", err))
		return
	}
	ctx = context.WithValue(ctx, fsctx.UploadPolicyCtx, *uploadPolicy)

	// 取得文件大小
	fileSize, err := strconv.ParseUint(c.Request.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		c.JSON(200, ErrorResponse(err))
		return
	}

	// 解码文件名和路径
	fileName, err := url.QueryUnescape(c.Request.Header.Get("X-Cr-FileName"))
	if err != nil {
		c.JSON(200, ErrorResponse(err))
		return
	}

	fileData := local.FileStream{
		MIMEType: c.Request.Header.Get("Content-Type"),
		File:     c.Request.Body,
		Name:     fileName,
		Size:     fileSize,
	}

	// 给文件系统分配钩子
	fs.Use("BeforeUpload", filesystem.HookSlaveUploadValidate)
	fs.Use("AfterUploadCanceled", filesystem.HookDeleteTempFile)
	fs.Use("AfterUpload", filesystem.SlaveAfterUpload)
	fs.Use("AfterValidateFailed", filesystem.HookDeleteTempFile)

	// 是否允许覆盖
	if c.Request.Header.Get("X-Cr-Overwrite") == "false" {
		ctx = context.WithValue(ctx, fsctx.DisableOverwrite, true)
	}

	// 执行上传
	err = fs.Upload(ctx, fileData)
	if err != nil {
		c.JSON(200, serializer.Err(serializer.CodeUploadFailed, err.Error(), err))
		return
	}

	c.JSON(200, serializer.Response{
		Code: 0,
	})
}

// SlaveDownload 从机文件下载,此请求返回的HTTP状态码不全为200
func SlaveDownload(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.SlaveDownloadService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.ServeFile(ctx, c, true)
		if res.Code != 0 {
			c.JSON(400, res)
		}
	} else {
		c.JSON(400, ErrorResponse(err))
	}
}

// SlavePreview 从机文件预览
func SlavePreview(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.SlaveDownloadService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.ServeFile(ctx, c, false)
		if res.Code != 0 {
			c.JSON(200, res)
		}
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// SlaveThumb 从机文件缩略图
func SlaveThumb(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.SlaveFileService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.Thumb(ctx, c)
		if res.Code != 0 {
			c.JSON(200, res)
		}
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// SlaveDelete 从机删除
func SlaveDelete(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.SlaveFilesService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Delete(ctx, c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// SlavePing 从机测试
func SlavePing(c *gin.Context) {
	var service admin.SlavePingService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Test()
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// SlaveList 从机列出文件
func SlaveList(c *gin.Context) {
	var service explorer.SlaveListService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.List(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// SlaveHeartbeat 接受主机心跳包
func SlaveHeartbeat(c *gin.Context) {
	var service serializer.NodePingReq
	if err := c.ShouldBindJSON(&service); err == nil {
		res := node.HandleMasterHeartbeat(&service)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// SlaveAria2Create 创建 Aria2 任务
func SlaveAria2Create(c *gin.Context) {
	var service serializer.SlaveAria2Call
	if err := c.ShouldBindJSON(&service); err == nil {
		res := aria2.Add(c, &service)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// SlaveAria2Status 查询从机 Aria2 任务状态
func SlaveAria2Status(c *gin.Context) {
	var service serializer.SlaveAria2Call
	if err := c.ShouldBindJSON(&service); err == nil {
		res := aria2.SlaveStatus(c, &service)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// SlaveCancelAria2Task 取消从机离线下载任务
func SlaveCancelAria2Task(c *gin.Context) {
	var service serializer.SlaveAria2Call
	if err := c.ShouldBindJSON(&service); err == nil {
		res := aria2.SlaveCancel(c, &service)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// SlaveSelectTask 从机选取离线下载文件
func SlaveSelectTask(c *gin.Context) {
	var service serializer.SlaveAria2Call
	if err := c.ShouldBindJSON(&service); err == nil {
		res := aria2.SlaveSelect(c, &service)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// SlaveCreateTransferTask 从机创建中转任务
func SlaveCreateTransferTask(c *gin.Context) {
	var service serializer.SlaveTransferReq
	if err := c.ShouldBindJSON(&service); err == nil {
		res := explorer.CreateTransferTask(c, &service)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// SlaveNotificationPush 处理从机发送的消息推送
func SlaveNotificationPush(c *gin.Context) {
	var service node.SlaveNotificationService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.HandleSlaveNotificationPush(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// SlaveGetOneDriveCredential 从机获取主机的OneDrive存储策略凭证
func SlaveGetOneDriveCredential(c *gin.Context) {
	var service node.OneDriveCredentialService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.Get(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// SlaveSelectTask 从机删除离线下载临时文件
func SlaveDeleteTempFile(c *gin.Context) {
	var service serializer.SlaveAria2Call
	if err := c.ShouldBindJSON(&service); err == nil {
		res := aria2.SlaveDeleteTemp(c, &service)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}
