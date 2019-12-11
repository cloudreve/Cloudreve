package explorer

import (
	"context"
	"github.com/HFO4/cloudreve/pkg/filesystem"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

// FileDownloadService 文件下载服务，path为文件完整路径
type FileDownloadService struct {
	Path string `uri:"path" binding:"required,min=1,max=65535"`
}

type FileAnonymousGetService struct {
	ID   uint   `uri:"id" binding:"required,min=1"`
	Name string `uri:"name" binding:"required"`
}

// Download 签名的匿名文件下载
func (service *FileAnonymousGetService) Download(ctx context.Context, c *gin.Context) serializer.Response {
	fs, err := filesystem.NewAnonymousFileSystem()
	if err != nil {
		return serializer.Err(serializer.CodeGroupNotAllowed, err.Error(), err)
	}

	// 查找文件
	err = fs.SetTargetFileByIDs([]uint{service.ID})
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	// 获取文件流
	rs, err := fs.GetDownloadContent(ctx, "")
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	// 发送文件
	http.ServeContent(c.Writer, c.Request, service.Name, fs.FileTarget[0].UpdatedAt, rs)

	// 检查是否需要关闭文件
	if fc, ok := rs.(io.Closer); ok {
		defer fc.Close()
	}

	return serializer.Response{
		Code: 0,
	}
}

// Download 文件下载
func (service *FileDownloadService) Download(ctx context.Context, c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}

	// 开始处理下载
	ctx = context.WithValue(ctx, fsctx.GinCtx, c)
	rs, err := fs.GetDownloadContent(ctx, service.Path)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	// 设置文件名
	c.Header("Content-Disposition", "attachment; filename=\""+fs.FileTarget[0].Name+"\"")
	// 发送文件
	http.ServeContent(c.Writer, c.Request, "", fs.FileTarget[0].UpdatedAt, rs)

	// 检查是否需要关闭文件
	if fc, ok := rs.(io.Closer); ok {
		defer fc.Close()
	}

	return serializer.Response{
		Code: 0,
	}
}
