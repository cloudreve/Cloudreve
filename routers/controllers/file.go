package controllers

import (
	"context"
	"fmt"
	"net/http"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/service/explorer"
	"github.com/gin-gonic/gin"
)

func DownloadArchive(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.ArchiveService
	if err := c.ShouldBindUri(&service); err == nil {
		service.DownloadArchived(ctx, c)
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

// Relocate 创建文件转移任务
func Relocate(c *gin.Context) {
	var service explorer.ItemRelocateService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.CreateRelocateTask(c)
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

// AnonymousPermLink Deprecated 文件签名后的永久链接
func AnonymousPermLinkDeprecated(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.FileAnonymousGetService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.Source(ctx, c)
		// 是否需要重定向
		if res.Code == -302 {
			c.Redirect(302, res.Data.(string))
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

// AnonymousPermLink 文件中转后的永久直链接
func AnonymousPermLink(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sourceLinkRaw, ok := c.Get("source_link")
	if !ok {
		c.JSON(200, serializer.Err(serializer.CodeFileNotFound, "", nil))
		return
	}

	sourceLink := sourceLinkRaw.(*model.SourceLink)

	service := &explorer.FileAnonymousGetService{
		ID:   sourceLink.FileID,
		Name: sourceLink.File.Name,
	}

	res := service.Source(ctx, c)
	// 是否需要重定向
	if res.Code == -302 {
		c.Redirect(302, res.Data.(string))
		return
	}

	// 是否有错误发生
	if res.Code != 0 {
		c.JSON(200, res)
	}

}

// GetSource 获取文件的外链地址
func GetSource(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.ItemIDService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Sources(ctx, c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
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
		c.JSON(200, serializer.Err(serializer.CodeFileNotFound, "", err))
		return
	}

	// 获取缩略图
	resp, err := fs.GetThumb(ctx, fileID.(uint))
	if err != nil {
		c.JSON(200, serializer.Err(serializer.CodeNotSet, "Failed to get thumbnail", err))
		return
	}

	if resp.Redirect {
		c.Header("Cache-Control", fmt.Sprintf("max-age=%d", resp.MaxAge))
		c.Redirect(http.StatusMovedPermanently, resp.URL)
		return
	}

	defer resp.Content.Close()
	http.ServeContent(c.Writer, c.Request, "thumb."+model.GetSettingByNameWithDefault("thumb_encode_method", "jpg"), fs.FileTarget[0].UpdatedAt, resp.Content)

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
			c.Redirect(302, res.Data.(string))
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
		res := service.CreateDocPreviewSession(ctx, c, true)
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

// FileUpload 本地策略文件上传
func FileUpload(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.UploadService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.LocalUpload(ctx, c)
		c.JSON(200, res)
		request.BlackHole(c.Request.Body)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// DeleteUploadSession 删除上传会话
func DeleteUploadSession(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.UploadSessionService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.Delete(ctx, c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// DeleteAllUploadSession 删除全部上传会话
func DeleteAllUploadSession(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res := explorer.DeleteAllUploadSession(ctx, c)
	c.JSON(200, res)
}

// GetUploadSession 创建上传会话
func GetUploadSession(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.CreateUploadSessionService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Create(ctx, c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// SearchFile 搜索文件
func SearchFile(c *gin.Context) {
	var service explorer.ItemSearchService
	if err := c.ShouldBindUri(&service); err != nil {
		c.JSON(200, ErrorResponse(err))
		return
	}

	if err := c.ShouldBindQuery(&service); err != nil {
		c.JSON(200, ErrorResponse(err))
		return
	}

	res := service.Search(c)
	c.JSON(200, res)
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
