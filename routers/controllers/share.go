package controllers

import (
	"context"
	"path"
	"strings"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/cloudreve/Cloudreve/v3/service/share"
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

// ListShare 列出分享
func ListShare(c *gin.Context) {
	var service share.ShareListService
	if err := c.ShouldBindQuery(&service); err == nil {
		res := service.List(c, CurrentUser(c))
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// SearchShare 搜索分享
func SearchShare(c *gin.Context) {
	var service share.ShareListService
	if err := c.ShouldBindQuery(&service); err == nil {
		res := service.Search(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// UpdateShare 更新分享属性
func UpdateShare(c *gin.Context) {
	var service share.ShareUpdateService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Update(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// DeleteShare 删除分享
func DeleteShare(c *gin.Context) {
	var service share.Service
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.Delete(c, CurrentUser(c))
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// GetShareDownload 创建分享下载会话
func GetShareDownload(c *gin.Context) {
	var service share.Service
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

	var service share.Service
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

	var service share.Service
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

// PreviewShareReadme 预览文本自述文件
func PreviewShareReadme(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service share.Service
	if err := c.ShouldBindQuery(&service); err == nil {
		// 自述文件名限制
		allowFileName := []string{"readme.txt", "readme.md"}
		fileName := strings.ToLower(path.Base(service.Path))
		if !util.ContainsString(allowFileName, fileName) {
			c.JSON(200, serializer.ParamErr("非README文件", nil))
		}

		// 必须是目录分享
		if shareCtx, ok := c.Get("share"); ok {
			if !shareCtx.(*model.Share).IsDir {
				c.JSON(200, serializer.ParamErr("此分享无自述文件", nil))
			}
		}

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
	var service share.Service
	if err := c.ShouldBindQuery(&service); err == nil {
		res := service.CreateDocPreviewSession(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// ListSharedFolder 列出分享的目录下的对象
func ListSharedFolder(c *gin.Context) {
	var service share.Service
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.List(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// ArchiveShare 打包要下载的分享
func ArchiveShare(c *gin.Context) {
	var service share.ArchiveService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Archive(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// ShareThumb 获取分享目录下文件的缩略图
func ShareThumb(c *gin.Context) {
	var service share.Service
	if err := c.ShouldBindQuery(&service); err == nil {
		res := service.Thumb(c)
		if res.Code >= 0 {
			c.JSON(200, res)
		}
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// GetUserShare 查看给定用户的分享
func GetUserShare(c *gin.Context) {
	var service share.ShareUserGetService
	if err := c.ShouldBindQuery(&service); err == nil {
		res := service.Get(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}
