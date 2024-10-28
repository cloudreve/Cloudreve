package explorer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/wopi"
	"github.com/gin-gonic/gin"
)

// SingleFileService 对单文件进行操作的五福，path为文件完整路径
type SingleFileService struct {
	Path string `uri:"path" json:"path" binding:"required,min=1,max=65535"`
}

// FileIDService 通过文件ID对文件进行操作的服务
type FileIDService struct {
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

// ArchiveService 文件流式打包下載服务
type ArchiveService struct {
	ID string `uri:"sessionID" binding:"required"`
}

// New 创建新文件
func (service *SingleFileService) Create(c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	// 上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 给文件系统分配钩子
	fs.Use("BeforeUpload", filesystem.HookValidateFile)
	fs.Use("AfterUpload", filesystem.GenericAfterUpload)

	// 上传空文件
	err = fs.Upload(ctx, &fsctx.FileStream{
		File:        ioutil.NopCloser(strings.NewReader("")),
		Size:        0,
		VirtualPath: path.Dir(service.Path),
		Name:        path.Base(service.Path),
	})
	if err != nil {
		return serializer.Err(serializer.CodeUploadFailed, err.Error(), err)
	}

	return serializer.Response{
		Code: 0,
	}
}

// List 列出从机上的文件
func (service *SlaveListService) List(c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewAnonymousFileSystem()
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	objects, err := fs.Handler.List(context.Background(), service.Path, service.Recursive)
	if err != nil {
		return serializer.Err(serializer.CodeIOFailed, "Cannot list files", err)
	}

	res, _ := json.Marshal(objects)
	return serializer.Response{Data: string(res)}
}

// DownloadArchived 通过预签名 URL 打包下载
func (service *ArchiveService) DownloadArchived(ctx context.Context, c *gin.Context) serializer.Response {
	userRaw, exist := cache.Get("archive_user_" + service.ID)
	if !exist {
		return serializer.Err(serializer.CodeNotFound, "Archive session not exist", nil)
	}
	user := userRaw.(model.User)

	// 创建文件系统
	fs, err := filesystem.NewFileSystem(&user)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	// 查找打包的临时文件
	archiveSession, exist := cache.Get("archive_" + service.ID)
	if !exist {
		return serializer.Err(serializer.CodeNotFound, "Archive session not exist", nil)
	}

	// 开始打包
	c.Header("Content-Disposition", "attachment;")
	c.Header("Content-Type", "application/zip")
	itemService := archiveSession.(ItemIDService)
	items := itemService.Raw()
	ctx = context.WithValue(ctx, fsctx.GinCtx, c)
	err = fs.Compress(ctx, c.Writer, items.Dirs, items.Items, true)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, "Failed to compress file", err)
	}

	return serializer.Response{
		Code: 0,
	}
}

// Download 签名的匿名文件下载
func (service *FileAnonymousGetService) Download(ctx context.Context, c *gin.Context) serializer.Response {
	fs, err := filesystem.NewAnonymousFileSystem()
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	// 查找文件
	err = fs.SetTargetFileByIDs([]uint{service.ID})
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	// 获取文件流
	rs, err := fs.GetDownloadContent(ctx, 0)
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

// Source 重定向到文件的有效原始链接
func (service *FileAnonymousGetService) Source(ctx context.Context, c *gin.Context) serializer.Response {
	fs, err := filesystem.NewAnonymousFileSystem()
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	// 查找文件
	err = fs.SetTargetFileByIDs([]uint{service.ID})
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	// 获取文件流
	ttl := int64(model.GetIntSetting("preview_timeout", 60))
	res, err := fs.SignURL(ctx, &fs.FileTarget[0], ttl, false)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	c.Header("Cache-Control", fmt.Sprintf("max-age=%d", ttl))
	return serializer.Response{
		Code: -302,
		Data: res,
	}
}

// CreateDocPreviewSession 创建DOC文件预览会话，返回预览地址
func (service *FileIDService) CreateDocPreviewSession(ctx context.Context, c *gin.Context, editable bool) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}
	defer fs.Recycle()

	// 获取对象id
	objectID, _ := c.Get("object_id")

	// 如果上下文中已有File对象，则重设目标
	if file, ok := ctx.Value(fsctx.FileModelCtx).(*model.File); ok {
		fs.SetTargetFile(&[]model.File{*file})
		objectID = uint(0)
	}

	// 如果上下文中已有Folder对象，则重设根目录
	if folder, ok := ctx.Value(fsctx.FolderModelCtx).(*model.Folder); ok {
		fs.Root = folder
		path := ctx.Value(fsctx.PathCtx).(string)
		err := fs.ResetFileIfNotExist(ctx, path)
		if err != nil {
			return serializer.Err(serializer.CodeNotFound, err.Error(), err)
		}
		objectID = uint(0)
	}

	// 获取文件临时下载地址
	downloadURL, err := fs.GetDownloadURL(ctx, objectID.(uint), "doc_preview_timeout")
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	// For newer version of Cloudreve - Local Policy
	// When do not use a cdn, the downloadURL withouts hosts, like "/api/v3/file/download/xxx"
	if strings.HasPrefix(downloadURL, "/") {
		downloadURI, err := url.Parse(downloadURL)
		if err != nil {
			return serializer.Err(serializer.CodeNotSet, err.Error(), err)
		}
		downloadURL = model.GetSiteURL().ResolveReference(downloadURI).String()
	}

	var resp serializer.DocPreviewSession

	// Use WOPI preview if available
	if model.IsTrueVal(model.GetSettingByName("wopi_enabled")) && wopi.Default != nil {
		maxSize := model.GetIntSetting("maxEditSize", 0)
		if maxSize > 0 && fs.FileTarget[0].Size > uint64(maxSize) {
			return serializer.Err(serializer.CodeFileTooLarge, "", nil)
		}

		action := wopi.ActionPreview
		if editable {
			action = wopi.ActionEdit
		}

		session, err := wopi.Default.NewSession(fs.FileTarget[0].UserID, &fs.FileTarget[0], action)
		if err != nil {
			return serializer.Err(serializer.CodeInternalSetting, "Failed to create WOPI session", err)
		}

		resp.URL = session.ActionURL.String()
		resp.AccessTokenTTL = session.AccessTokenTTL
		resp.AccessToken = session.AccessToken
		return serializer.Response{
			Code: 0,
			Data: resp,
		}
	}

	// 生成最终的预览器地址
	srcB64 := base64.StdEncoding.EncodeToString([]byte(downloadURL))
	srcEncoded := url.QueryEscape(downloadURL)
	srcB64Encoded := url.QueryEscape(srcB64)
	resp.URL = util.Replace(map[string]string{
		"{$src}":    srcEncoded,
		"{$srcB64}": srcB64Encoded,
		"{$name}":   url.QueryEscape(fs.FileTarget[0].Name),
	}, model.GetSettingByName("office_preview_service"))

	return serializer.Response{
		Code: 0,
		Data: resp,
	}
}

// CreateDownloadSession 创建下载会话，获取下载URL
func (service *FileIDService) CreateDownloadSession(ctx context.Context, c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	// 获取对象id
	objectID, _ := c.Get("object_id")

	// 获取下载地址
	downloadURL, err := fs.GetDownloadURL(ctx, objectID.(uint), "download_timeout")
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
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	// 查找打包的临时文件
	file, exist := cache.Get("download_" + service.ID)
	if !exist {
		return serializer.Err(serializer.CodeNotFound, "Download session not exist", nil)
	}
	fs.FileTarget = []model.File{file.(model.File)}

	// 开始处理下载
	ctx = context.WithValue(ctx, fsctx.GinCtx, c)
	rs, err := fs.GetDownloadContent(ctx, 0)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}
	defer rs.Close()

	// 设置文件名
	c.Header("Content-Disposition", "attachment; filename=\""+url.PathEscape(fs.FileTarget[0].Name)+"\"")

	if fs.User.Group.OptionsSerialized.OneTimeDownload {
		// 清理资源，删除临时文件
		_ = cache.Deletes([]string{service.ID}, "download_")
	}

	// 发送文件
	http.ServeContent(c.Writer, c.Request, fs.FileTarget[0].Name, fs.FileTarget[0].UpdatedAt, rs)

	return serializer.Response{
		Code: 0,
	}
}

// PreviewContent 预览文件，需要登录会话, isText - 是否为文本文件，文本文件会
// 强制经由服务端中转
func (service *FileIDService) PreviewContent(ctx context.Context, c *gin.Context, isText bool) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	// 获取对象id
	objectID, _ := c.Get("object_id")

	// 如果上下文中已有File对象，则重设目标
	if file, ok := ctx.Value(fsctx.FileModelCtx).(*model.File); ok {
		fs.SetTargetFile(&[]model.File{*file})
		objectID = uint(0)
	}

	// 如果上下文中已有Folder对象，则重设根目录
	if folder, ok := ctx.Value(fsctx.FolderModelCtx).(*model.Folder); ok {
		fs.Root = folder
		path := ctx.Value(fsctx.PathCtx).(string)
		err := fs.ResetFileIfNotExist(ctx, path)
		if err != nil {
			return serializer.Err(serializer.CodeFileNotFound, err.Error(), err)
		}
		objectID = uint(0)
	}

	// 获取文件预览响应
	resp, err := fs.Preview(ctx, objectID.(uint), isText)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	// 重定向到文件源
	if resp.Redirect {
		c.Header("Cache-Control", fmt.Sprintf("max-age=%d", resp.MaxAge))
		return serializer.Response{
			Code: -301,
			Data: resp.URL,
		}
	}

	// 直接返回文件内容
	defer resp.Content.Close()

	if isText {
		c.Header("Cache-Control", "no-cache")
	}

	http.ServeContent(c.Writer, c.Request, fs.FileTarget[0].Name, fs.FileTarget[0].UpdatedAt, resp.Content)

	return serializer.Response{
		Code: 0,
	}
}

// PutContent 更新文件内容
func (service *FileIDService) PutContent(ctx context.Context, c *gin.Context) serializer.Response {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 取得文件大小
	fileSize, err := strconv.ParseUint(c.Request.Header.Get("Content-Length"), 10, 64)
	if err != nil {

		return serializer.ParamErr("Invalid content-length value", err)
	}

	// 计算hash
	file, hook := filesystem.NewChecksumFileStreamAndAfterUploadHook(c.Request.Body)

	fileData := fsctx.FileStream{
		MimeType: c.Request.Header.Get("Content-Type"),
		File:     file,
		Size:     fileSize,
		Mode:     fsctx.Overwrite,
	}

	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	uploadCtx := context.WithValue(ctx, fsctx.GinCtx, c)

	// 取得现有文件
	fileID, _ := c.Get("object_id")
	originFile, _ := model.GetFilesByIDs([]uint{fileID.(uint)}, fs.User.ID)
	if len(originFile) == 0 {
		return serializer.Err(serializer.CodeFileNotFound, "", nil)
	}
	fileData.Name = originFile[0].Name

	// 检查此文件是否有软链接
	fileList, err := model.RemoveFilesWithSoftLinks([]model.File{originFile[0]})
	if err == nil && len(fileList) == 0 {
		// 如果包含软连接，应重新生成新文件副本，并更新source_name
		originFile[0].SourceName = fs.GenerateSavePath(uploadCtx, &fileData)
		fileData.Mode &= ^fsctx.Overwrite
		fs.Use("AfterUpload", filesystem.HookUpdateSourceName)
		fs.Use("AfterUploadCanceled", filesystem.HookUpdateSourceName)
		fs.Use("AfterUploadCanceled", filesystem.HookCleanFileContent)
		fs.Use("AfterUploadCanceled", filesystem.HookClearFileSize)
		fs.Use("AfterValidateFailed", filesystem.HookUpdateSourceName)
		fs.Use("AfterValidateFailed", filesystem.HookCleanFileContent)
		fs.Use("AfterValidateFailed", filesystem.HookClearFileSize)
	}

	// 给文件系统分配钩子
	fs.Use("BeforeUpload", filesystem.HookResetPolicy)
	fs.Use("BeforeUpload", filesystem.HookValidateFile)
	fs.Use("BeforeUpload", filesystem.HookValidateCapacityDiff)
	fs.Use("AfterUpload", filesystem.GenericAfterUpdate)
	fs.Use("AfterUpload", hook)

	// 执行上传
	uploadCtx = context.WithValue(uploadCtx, fsctx.FileModelCtx, originFile[0])
	err = fs.Upload(uploadCtx, &fileData)
	if err != nil {
		return serializer.Err(serializer.CodeUploadFailed, err.Error(), err)
	}

	return serializer.Response{
		Code: 0,
	}
}

// Sources 批量获取对象的外链
func (s *ItemIDService) Sources(ctx context.Context, c *gin.Context) serializer.Response {
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	if len(s.Raw().Items) > fs.User.Group.OptionsSerialized.SourceBatchSize {
		return serializer.Err(serializer.CodeBatchSourceSize, "", err)
	}

	res := make([]serializer.Sources, 0, len(s.Raw().Items))
	files, err := model.GetFilesByIDs(s.Raw().Items, fs.User.ID)
	if err != nil || len(files) == 0 {
		return serializer.Err(serializer.CodeFileNotFound, "", err)
	}

	getSourceFunc := func(file model.File) (string, error) {
		fs.FileTarget = []model.File{file}
		return fs.GetSource(ctx, file.ID)
	}

	// Create redirected source link if needed
	if fs.User.Group.OptionsSerialized.RedirectedSource {
		getSourceFunc = func(file model.File) (string, error) {
			source, err := file.CreateOrGetSourceLink()
			if err != nil {
				return "", err
			}

			sourceLinkURL, err := source.Link()
			if err != nil {
				return "", err
			}

			return sourceLinkURL, nil
		}
	}

	for _, file := range files {
		sourceURL, err := getSourceFunc(file)
		current := serializer.Sources{
			URL:    sourceURL,
			Name:   file.Name,
			Parent: file.FolderID,
		}

		if err != nil {
			current.Error = err.Error()
		}

		res = append(res, current)
	}

	return serializer.Response{
		Code: 0,
		Data: res,
	}
}
