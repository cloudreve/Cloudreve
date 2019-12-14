package explorer

import (
	"context"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/cache"
	"github.com/HFO4/cloudreve/pkg/filesystem"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// FileDownloadCreateService 文件下载会话创建服务，path为文件完整路径
type FileDownloadCreateService struct {
	Path string `uri:"path" binding:"required,min=1,max=65535"`
}

// FileAnonymousGetService 匿名（外链）获取文件服务
type FileAnonymousGetService struct {
	ID   uint   `uri:"id" binding:"required,min=1"`
	Name string `uri:"name" binding:"required"`
}

// DownloadService 文件下載服务
type DownloadService struct {
	ID string `uri:"id" binding:"required"`
}

// DownloadArchived 下載已打包的多文件
func (service *DownloadService) DownloadArchived(ctx context.Context, c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}

	// 查找打包的临时文件
	zipPath, exist := cache.Get("archive_" + service.ID)
	if !exist {
		return serializer.Err(404, "归档文件不存在", nil)
	}

	// 获取文件流
	rs, err := fs.GetPhysicalFileContent(ctx, zipPath.(string))
	defer rs.Close()
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	if fs.User.Group.OptionsSerialized.OneTimeDownloadEnabled {
		// 清理资源，删除临时文件
		_ = cache.Deletes([]string{service.ID}, "archive_")
	}

	c.Header("Content-Disposition", "attachment;")
	c.Header("Content-Type", "application/zip")
	http.ServeContent(c.Writer, c.Request, "", time.Now(), rs)

	return serializer.Response{
		Code: 0,
	}

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
	defer rs.Close()
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	// 发送文件
	http.ServeContent(c.Writer, c.Request, service.Name, fs.FileTarget[0].UpdatedAt, rs)

	return serializer.Response{
		Code: 0,
	}
}

// CreateDownloadSession 创建下载会话，获取下载URL
func (service *FileDownloadCreateService) CreateDownloadSession(ctx context.Context, c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}

	// 获取下载地址
	downloadURL, err := fs.GetDownloadURL(ctx, service.Path)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	return serializer.Response{
		Code: 0,
		Data: downloadURL,
	}
}

// Download 通过签名URL的文件下载，无需登录
func (service *DownloadService) Download(ctx context.Context, c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}

	// 查找打包的临时文件
	file, exist := cache.Get("download_" + service.ID)
	if !exist {
		return serializer.Err(404, "文件下载会话不存在", nil)
	}
	fs.FileTarget = []model.File{file.(model.File)}

	// 开始处理下载
	ctx = context.WithValue(ctx, fsctx.GinCtx, c)
	rs, err := fs.GetDownloadContent(ctx, "")
	defer rs.Close()
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	// 设置文件名
	c.Header("Content-Disposition", "attachment; filename=\""+fs.FileTarget[0].Name+"\"")

	if fs.User.Group.OptionsSerialized.OneTimeDownloadEnabled {
		// 清理资源，删除临时文件
		_ = cache.Deletes([]string{service.ID}, "download_")
	}

	// 发送文件
	http.ServeContent(c.Writer, c.Request, "", fs.FileTarget[0].UpdatedAt, rs)

	return serializer.Response{
		Code: 0,
	}
}

// PreviewContent 预览文件，需要登录会话
func (service *FileDownloadCreateService) PreviewContent(ctx context.Context, c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}

	// 获取文件流
	rs, err := fs.GetDownloadContent(ctx, service.Path)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}
	defer rs.Close()

	http.ServeContent(c.Writer, c.Request, fs.FileTarget[0].Name, fs.FileTarget[0].UpdatedAt, rs)

	return serializer.Response{
		Code: 0,
	}
}
