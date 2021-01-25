package controllers

import "C"
import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/local"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/service/explorer"
	"github.com/gin-gonic/gin"
)

func DownloadArchive(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.DownloadService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.DownloadArchived(ctx, c)
		if res.Code != 0 {
			c.JSON(200, res)
		}
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

func Archive(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.ItemIDService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Archive(ctx, c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// Compress 创建文件压缩任务
func Compress(c *gin.Context) {
	var service explorer.ItemCompressService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.CreateCompressTask(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// Decompress 创建文件解压缩任务
func Decompress(c *gin.Context) {
	var service explorer.ItemDecompressService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.CreateDecompressTask(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// AnonymousGetContent 匿名获取文件资源
func AnonymousGetContent(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.FileAnonymousGetService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.Download(ctx, c)
		if res.Code != 0 {
			c.JSON(200, res)
		}
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// GetSource 获取文件的外链地址
func GetSource(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		c.JSON(200, serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err))
		return
	}
	defer fs.Recycle()

	// 获取文件ID
	fileID, ok := c.Get("object_id")
	if !ok {
		c.JSON(200, serializer.ParamErr("文件不存在", err))
		return
	}

	sourceURL, err := fs.GetSource(ctx, fileID.(uint))
	if err != nil {
		c.JSON(200, serializer.Err(serializer.CodeNotSet, err.Error(), err))
		return
	}

	c.JSON(200, serializer.Response{
		Code: 0,
		Data: struct {
			URL string `json:"url"`
		}{URL: sourceURL},
	})

}

// Thumb 获取文件缩略图
func Thumb(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		c.JSON(200, serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err))
		return
	}
	defer fs.Recycle()

	// 获取文件ID
	fileID, ok := c.Get("object_id")
	if !ok {
		c.JSON(200, serializer.ParamErr("文件不存在", err))
		return
	}

	// 获取缩略图
	resp, err := fs.GetThumb(ctx, fileID.(uint))
	if err != nil {
		c.JSON(200, serializer.Err(serializer.CodeNotSet, "无法获取缩略图", err))
		return
	}

	if resp.Redirect {
		c.Header("Cache-Control", fmt.Sprintf("max-age=%d", resp.MaxAge))
		c.Redirect(http.StatusMovedPermanently, resp.URL)
		return
	}

	defer resp.Content.Close()
	http.ServeContent(c.Writer, c.Request, "thumb.png", fs.FileTarget[0].UpdatedAt, resp.Content)

}

// Preview 预览文件
func Preview(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.FileIDService
	if err := c.ShouldBindUri(&service); err == nil {
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

// PreviewText 预览文本文件
func PreviewText(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.FileIDService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.PreviewContent(ctx, c, true)
		// 是否有错误发生
		if res.Code != 0 {
			c.JSON(200, res)
		}
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// GetDocPreview 获取DOC文件预览地址
func GetDocPreview(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.FileIDService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.CreateDocPreviewSession(ctx, c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// CreateDownloadSession 创建文件下载会话
func CreateDownloadSession(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.FileIDService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.CreateDownloadSession(ctx, c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// Download 文件下载
func Download(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.DownloadService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.Download(ctx, c)
		if res.Code != 0 {
			c.JSON(200, res)
		}
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// PutContent 更新文件内容
func PutContent(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.FileIDService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.PutContent(ctx, c)
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

	// 取得文件大小
	fileSize, err := strconv.ParseUint(c.Request.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		c.JSON(200, ErrorResponse(err))
		return
	}

	// 非可用策略时拒绝上传
	if user, ok := c.Get("user"); ok && !user.(*model.User).Policy.IsTransitUpload(fileSize) {
		request.BlackHole(c.Request.Body)
		c.JSON(200, serializer.Err(serializer.CodePolicyNotAllowed, "当前存储策略无法使用", nil))
		return
	}

	// 解码文件名和路径
	fileName, err := url.QueryUnescape(c.Request.Header.Get("X-FileName"))
	filePath, err := url.QueryUnescape(c.Request.Header.Get("X-Path"))
	if err != nil {
		c.JSON(200, ErrorResponse(err))
		return
	}

	fileData := local.FileStream{
		MIMEType:    c.Request.Header.Get("Content-Type"),
		File:        c.Request.Body,
		Size:        fileSize,
		Name:        fileName,
		VirtualPath: filePath,
	}

	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		c.JSON(200, serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err))
		return
	}

	// 给文件系统分配钩子
	fs.Use("BeforeUpload", filesystem.HookValidateFile)
	fs.Use("BeforeUpload", filesystem.HookValidateCapacity)
	fs.Use("AfterUploadCanceled", filesystem.HookDeleteTempFile)
	fs.Use("AfterUploadCanceled", filesystem.HookGiveBackCapacity)
	fs.Use("AfterUpload", filesystem.GenericAfterUpload)
	fs.Use("AfterValidateFailed", filesystem.HookDeleteTempFile)
	fs.Use("AfterValidateFailed", filesystem.HookGiveBackCapacity)
	fs.Use("AfterUploadFailed", filesystem.HookGiveBackCapacity)

	// 执行上传
	ctx = context.WithValue(ctx, fsctx.ValidateCapacityOnceCtx, &sync.Once{})
	uploadCtx := context.WithValue(ctx, fsctx.GinCtx, c)
	err = fs.Upload(uploadCtx, fileData)
	if err != nil {
		c.JSON(200, serializer.Err(serializer.CodeUploadFailed, err.Error(), err))
		return
	}

	c.JSON(200, serializer.Response{
		Code: 0,
	})
}

// GetUploadCredential 获取上传凭证
func GetUploadCredential(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.UploadCredentialService
	if err := c.ShouldBindQuery(&service); err == nil {
		res := service.Get(ctx, c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// SearchFile 搜索文件
func SearchFile(c *gin.Context) {
	var service explorer.ItemSearchService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.Search(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// CreateFile 创建空白文件
func CreateFile(c *gin.Context) {
	var service explorer.SingleFileService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Create(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}
