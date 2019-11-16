package file

import (
	"cloudreve/models"
	"cloudreve/pkg/filesystem"
	"cloudreve/pkg/filesystem/local"
	"cloudreve/pkg/serializer"
	"github.com/gin-gonic/gin"
	"mime/multipart"
)

// UploadService 本地策略文件上传服务
type UploadService struct {
	Name string                `form:"name" binding:"required,lte=255"`
	Path string                `form:"path" binding:"lte=65536"`
	File *multipart.FileHeader `form:"file" binding:"required"`
}

// Upload 处理本地策略小文件上传
func (service *UploadService) Upload(c *gin.Context) serializer.Response {
	// TODO 检查文件大小是否小于分片最大大小
	file, err := service.File.Open()
	if err != nil {
		return serializer.Err(serializer.CodeIOFailed, "无法打开上传数据", err)
	}

	fileData := local.FileData{
		MIMEType: service.File.Header.Get("Content-Type"),
		File:     file,
		Size:     uint64(service.File.Size),
		Name:     service.Name,
	}
	user, _ := c.Get("user")
	fs := filesystem.FileSystem{
		BeforeUpload: filesystem.GenericBeforeUpload,
		User:         user.(*model.User),
	}

	err = fs.Upload(fileData)
	if err != nil {
		return serializer.Err(serializer.CodeUploadFailed, err.Error(), err)
	}

	return serializer.Response{
		Code: 0,
		Msg:  "Pong",
	}
}
