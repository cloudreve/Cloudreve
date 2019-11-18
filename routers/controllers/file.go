package controllers

import (
	"context"
	"github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem"
	"github.com/HFO4/cloudreve/pkg/filesystem/local"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/service/file"
	"github.com/gin-gonic/gin"
	"strconv"
)

// FileUpload 本地策略文件上传
func FileUpload(c *gin.Context) {

	// 非本地策略时拒绝上传
	if user, ok := c.Get("user"); ok && user.(*model.User).Policy.Type != "local" {
		c.JSON(200, serializer.Err(serializer.CodePolicyNotAllowed, "当前上传策略无法使用", nil))
		return
	}

	// 建立上下文
	ctx, cancel := context.WithCancel(context.Background())

	var service file.UploadService
	defer cancel()

	if err := c.ShouldBind(&service); err == nil {
		res := service.Upload(ctx, c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// FileUploadStream 本地策略流式上传
func FileUploadStream(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 非本地策略时拒绝上传
	if user, ok := c.Get("user"); ok && user.(*model.User).Policy.Type != "local" {
		c.JSON(200, serializer.Err(serializer.CodePolicyNotAllowed, "当前上传策略无法使用", nil))
		return
	}

	// 取得文件大小
	fileSize, err := strconv.ParseUint(c.Request.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		c.JSON(200, ErrorResponse(err))
		return
	}

	fileData := local.FileStream{
		MIMEType: c.Request.Header.Get("Content-Type"),
		File:     c.Request.Body,
		Size:     fileSize,
		Name:     c.Request.Header.Get("X-FileName"),
	}
	user, _ := c.Get("user")

	// 创建文件系统
	fs, err := filesystem.NewFileSystem(user.(*model.User))
	if err != nil {
		c.JSON(200, serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err))
		return
	}

	// 给文件系统分配钩子
	fs.BeforeUpload = filesystem.GenericBeforeUpload
	fs.AfterUploadCanceled = filesystem.GenericAfterUploadCanceled
	fs.AfterUpload = filesystem.GenericAfterUpload
	fs.AfterValidateFailed = filesystem.GenericAfterUploadCanceled

	// 执行上传
	uploadCtx := context.WithValue(ctx, filesystem.GinCtx, c)
	err = fs.Upload(uploadCtx, fileData)
	if err != nil {
		c.JSON(200, serializer.Err(serializer.CodeUploadFailed, err.Error(), err))
		return
	}

	c.JSON(200, serializer.Response{
		Code: 0,
		Msg:  "Pong",
	})
}
