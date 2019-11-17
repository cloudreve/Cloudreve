package file

import (
	"context"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem"
	"github.com/HFO4/cloudreve/pkg/filesystem/local"
	"github.com/HFO4/cloudreve/pkg/serializer"
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
func (service *UploadService) Upload(ctx context.Context, c *gin.Context) serializer.Response {
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

	// 创建文件系统
	fs, err := filesystem.NewFileSystem(user.(*model.User))
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}

	// 给文件系统分配钩子
	fs.BeforeUpload = filesystem.GenericBeforeUpload

	// 执行上传
	err = fs.Upload(ctx, fileData)
	if err != nil {
		return serializer.Err(serializer.CodeUploadFailed, err.Error(), err)
	}

	return serializer.Response{
		Code: 0,
		Msg:  "Pong",
	}
}
