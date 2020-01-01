package controllers

import (
	"context"
	"github.com/HFO4/cloudreve/pkg/filesystem"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/filesystem/local"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/service/explorer"
	"github.com/gin-gonic/gin"
	"net/url"
	"strconv"
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
	fs.Handler = local.Handler{}

	// 从请求中取得上传策略
	uploadPolicyRaw := c.GetHeader("X-Policy")
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
	fileName, err := url.QueryUnescape(c.Request.Header.Get("X-FileName"))
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

// SlaveDownload 从机文件下载
func SlaveDownload(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.SlaveDownloadService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.ServeFile(ctx, c, true)
		if res.Code != 0 {
			c.JSON(200, res)
		}
	} else {
		c.JSON(200, ErrorResponse(err))
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
