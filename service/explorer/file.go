package explorer

import (
	"context"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/cache"
	"github.com/HFO4/cloudreve/pkg/filesystem"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/filesystem/local"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"
)

// SingleFileService 对单文件进行操作的五福，path为文件完整路径
type SingleFileService struct {
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
	defer fs.Recycle()

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
	defer fs.Recycle()

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

// CreateDocPreviewSession 创建DOC文件预览会话，返回预览地址
func (service *SingleFileService) CreateDocPreviewSession(ctx context.Context, c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}
	defer fs.Recycle()

	// 获取文件临时下载地址
	downloadURL, err := fs.GetDownloadURL(ctx, service.Path, "doc_preview_timeout")
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	// 生成最终的预览器地址
	// TODO 从配置文件中读取
	viewerBase, _ := url.Parse("https://view.officeapps.live.com/op/view.aspx")
	params := viewerBase.Query()
	params.Set("src", downloadURL)
	viewerBase.RawQuery = params.Encode()

	return serializer.Response{
		Code: 0,
		Data: viewerBase.String(),
	}
}

// CreateDownloadSession 创建下载会话，获取下载URL
func (service *SingleFileService) CreateDownloadSession(ctx context.Context, c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}
	defer fs.Recycle()

	// 获取下载地址
	downloadURL, err := fs.GetDownloadURL(ctx, service.Path, "download_timeout")
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
	defer fs.Recycle()

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
	c.Header("Content-Disposition", "attachment; filename=\""+url.PathEscape(fs.FileTarget[0].Name)+"\"")

	if fs.User.Group.OptionsSerialized.OneTimeDownloadEnabled {
		// 清理资源，删除临时文件
		_ = cache.Deletes([]string{service.ID}, "download_")
	}

	// 发送文件
	http.ServeContent(c.Writer, c.Request, fs.FileTarget[0].Name, fs.FileTarget[0].UpdatedAt, rs)

	return serializer.Response{
		Code: 0,
	}
}

// PreviewContent 预览文件，需要登录会话
func (service *SingleFileService) PreviewContent(ctx context.Context, c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}
	defer fs.Recycle()

	// 获取文件预览响应
	resp, err := fs.Preview(ctx, service.Path)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	// 重定向到文件源
	if resp.Redirect {
		return serializer.Response{
			Code: -301,
			Data: resp.URL,
		}
	}

	// 直接返回文件内容
	defer resp.Content.Close()

	http.ServeContent(c.Writer, c.Request, fs.FileTarget[0].Name, fs.FileTarget[0].UpdatedAt, resp.Content)

	return serializer.Response{
		Code: 0,
	}
}

// PutContent 更新文件内容
func (service *SingleFileService) PutContent(ctx context.Context, c *gin.Context) serializer.Response {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 取得文件大小
	fileSize, err := strconv.ParseUint(c.Request.Header.Get("Content-Length"), 10, 64)
	if err != nil {

		return serializer.ParamErr("无法解析文件尺寸", err)
	}

	fileData := local.FileStream{
		MIMEType:    c.Request.Header.Get("Content-Type"),
		File:        c.Request.Body,
		Size:        fileSize,
		Name:        path.Base(service.Path),
		VirtualPath: path.Dir(service.Path),
	}

	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}
	defer fs.Recycle()

	// 取得现有文件
	exist, originFile := fs.IsFileExist(service.Path)
	if !exist {
		return serializer.Err(404, "文件不存在", nil)
	}

	// 给文件系统分配钩子
	fs.Use("BeforeUpload", filesystem.HookValidateFile)
	fs.Use("BeforeUpload", filesystem.HookResetPolicy)
	fs.Use("BeforeUpload", filesystem.HookChangeCapacity)
	fs.Use("AfterUploadCanceled", filesystem.HookCleanFileContent)
	fs.Use("AfterUploadCanceled", filesystem.HookClearFileSize)
	fs.Use("AfterUploadCanceled", filesystem.HookGiveBackCapacity)
	fs.Use("AfterUpload", filesystem.GenericAfterUpdate)
	fs.Use("AfterValidateFailed", filesystem.HookCleanFileContent)
	fs.Use("AfterValidateFailed", filesystem.HookClearFileSize)
	fs.Use("AfterValidateFailed", filesystem.HookGiveBackCapacity)

	// 执行上传
	uploadCtx := context.WithValue(ctx, fsctx.GinCtx, c)
	uploadCtx = context.WithValue(uploadCtx, fsctx.FileModelCtx, *originFile)
	err = fs.Upload(uploadCtx, fileData)
	if err != nil {
		return serializer.Err(serializer.CodeUploadFailed, err.Error(), err)
	}

	return serializer.Response{
		Code: 0,
	}
}
