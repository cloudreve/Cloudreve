package explorer

import (
	"context"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/local"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-gonic/gin"
)

// CreateUploadSessionService 获取上传凭证服务
type CreateUploadSessionService struct {
	Path         string `json:"path" binding:"required"`
	Size         uint64 `json:"size" binding:"min=0"`
	Name         string `json:"name" binding:"required"`
	PolicyID     string `json:"policy_id" binding:"required"`
	LastModified int64  `json:"last_modified"`
	MimeType     string `json:"mime_type"`
}

// Create 创建新的上传会话
func (service *CreateUploadSessionService) Create(ctx context.Context, c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}

	// 取得存储策略的ID
	rawID, err := hashid.DecodeHashID(service.PolicyID, hashid.PolicyID)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotExist, "", err)
	}

	// 分配并检查存储策略
	if err := fs.SetPolicyFromPreference(rawID); err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, "", err)
	}

	if fs.Policy.ID != rawID {
		return serializer.Err(serializer.CodePolicyChanged, "", nil)
	}

	file := &fsctx.FileStream{
		Size:        service.Size,
		Name:        service.Name,
		VirtualPath: service.Path,
		File:        ioutil.NopCloser(strings.NewReader("")),
		MimeType:    service.MimeType,
	}
	if service.LastModified > 0 {
		lastModified := time.UnixMilli(service.LastModified)
		file.LastModified = &lastModified
	}
	credential, err := fs.CreateUploadSession(ctx, file)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	return serializer.Response{
		Code: 0,
		Data: credential,
	}
}

// UploadService 本机及从机策略上传服务
type UploadService struct {
	ID    string `uri:"sessionId" binding:"required"`
	Index int    `uri:"index" form:"index" binding:"min=0"`
}

// LocalUpload 处理本机文件分片上传
func (service *UploadService) LocalUpload(ctx context.Context, c *gin.Context) serializer.Response {
	uploadSessionRaw, ok := cache.Get(filesystem.UploadSessionCachePrefix + service.ID)
	if !ok {
		return serializer.Err(serializer.CodeUploadSessionExpired, "", nil)
	}

	uploadSession := uploadSessionRaw.(serializer.UploadSession)

	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}

	if uploadSession.UID != fs.User.ID {
		return serializer.Err(serializer.CodeUploadSessionExpired, "", nil)
	}

	// 查找上传会话创建的占位文件
	file, err := model.GetFilesByUploadSession(service.ID, fs.User.ID)
	if err != nil {
		return serializer.Err(serializer.CodeUploadSessionExpired, "", err)
	}

	// 重设 fs 存储策略
	if !uploadSession.Policy.IsTransitUpload(uploadSession.Size) {
		return serializer.Err(serializer.CodePolicyNotAllowed, "", err)
	}

	fs.Policy = &uploadSession.Policy
	if err := fs.DispatchHandler(); err != nil {
		return serializer.Err(serializer.CodePolicyNotExist, "", err)
	}

	expectedSizeStart := file.Size
	actualSizeStart := uint64(service.Index) * uploadSession.Policy.OptionsSerialized.ChunkSize
	if uploadSession.Policy.OptionsSerialized.ChunkSize == 0 && service.Index > 0 {
		return serializer.Err(serializer.CodeInvalidChunkIndex, "Chunk index cannot be greater than 0", nil)
	}

	if expectedSizeStart < actualSizeStart {
		return serializer.Err(serializer.CodeInvalidChunkIndex, "Chunk must be uploaded in order", nil)
	}

	if expectedSizeStart > actualSizeStart {
		util.Log().Info("Trying to overwrite chunk[%d] Start=%d", service.Index, actualSizeStart)
	}

	return processChunkUpload(ctx, c, fs, &uploadSession, service.Index, file, fsctx.Append)
}

// SlaveUpload 处理从机文件分片上传
func (service *UploadService) SlaveUpload(ctx context.Context, c *gin.Context) serializer.Response {
	uploadSessionRaw, ok := cache.Get(filesystem.UploadSessionCachePrefix + service.ID)
	if !ok {
		return serializer.Err(serializer.CodeUploadSessionExpired, "", nil)
	}

	uploadSession := uploadSessionRaw.(serializer.UploadSession)

	fs, err := filesystem.NewAnonymousFileSystem()
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}

	fs.Handler = local.Driver{}

	// 解析需要的参数
	service.Index, _ = strconv.Atoi(c.Query("chunk"))
	mode := fsctx.Append
	if c.GetHeader(auth.CrHeaderPrefix+"Overwrite") == "true" {
		mode |= fsctx.Overwrite
	}

	return processChunkUpload(ctx, c, fs, &uploadSession, service.Index, nil, mode)
}

func processChunkUpload(ctx context.Context, c *gin.Context, fs *filesystem.FileSystem, session *serializer.UploadSession, index int, file *model.File, mode fsctx.WriteMode) serializer.Response {
	// 取得并校验文件大小是否符合分片要求
	chunkSize := session.Policy.OptionsSerialized.ChunkSize
	isLastChunk := session.Policy.OptionsSerialized.ChunkSize == 0 || uint64(index+1)*chunkSize >= session.Size
	expectedLength := chunkSize
	if isLastChunk {
		expectedLength = session.Size - uint64(index)*chunkSize
	}

	fileSize, err := strconv.ParseUint(c.Request.Header.Get("Content-Length"), 10, 64)
	if err != nil || (expectedLength != fileSize) {
		return serializer.Err(
			serializer.CodeInvalidContentLength,
			fmt.Sprintf("Invalid Content-Length (expected: %d)", expectedLength),
			err,
		)
	}

	// 非首个分片时需要允许覆盖
	if index > 0 {
		mode |= fsctx.Overwrite
	}

	fileData := fsctx.FileStream{
		MimeType:     c.Request.Header.Get("Content-Type"),
		File:         c.Request.Body,
		Size:         fileSize,
		Name:         session.Name,
		VirtualPath:  session.VirtualPath,
		SavePath:     session.SavePath,
		Mode:         mode,
		AppendStart:  chunkSize * uint64(index),
		Model:        file,
		LastModified: session.LastModified,
	}

	// 给文件系统分配钩子
	fs.Use("AfterUploadCanceled", filesystem.HookTruncateFileTo(fileData.AppendStart))
	fs.Use("AfterValidateFailed", filesystem.HookTruncateFileTo(fileData.AppendStart))

	if file != nil {
		fs.Use("BeforeUpload", filesystem.HookValidateCapacity)
		fs.Use("AfterUpload", filesystem.HookChunkUploaded)
		fs.Use("AfterValidateFailed", filesystem.HookChunkUploadFailed)
		if isLastChunk {
			fs.Use("AfterUpload", filesystem.HookPopPlaceholderToFile(""))
			fs.Use("AfterUpload", filesystem.HookDeleteUploadSession(session.Key))
		}
	} else {
		if isLastChunk {
			fs.Use("AfterUpload", filesystem.SlaveAfterUpload(session))
			fs.Use("AfterUpload", filesystem.HookDeleteUploadSession(session.Key))
		}
	}

	// 执行上传
	uploadCtx := context.WithValue(ctx, fsctx.GinCtx, c)
	err = fs.Upload(uploadCtx, &fileData)
	if err != nil {
		return serializer.Err(serializer.CodeUploadFailed, err.Error(), err)
	}

	return serializer.Response{}
}

// UploadSessionService 上传会话服务
type UploadSessionService struct {
	ID string `uri:"sessionId" binding:"required"`
}

// Delete 删除指定上传会话
func (service *UploadSessionService) Delete(ctx context.Context, c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	// 查找需要删除的上传会话的占位文件
	file, err := model.GetFilesByUploadSession(service.ID, fs.User.ID)
	if err != nil {
		return serializer.Err(serializer.CodeUploadSessionExpired, "", err)
	}

	// 删除文件
	if err := fs.Delete(ctx, []uint{}, []uint{file.ID}, false, false); err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "Failed to delete upload session", err)
	}

	return serializer.Response{}
}

// SlaveDelete 从机删除指定上传会话
func (service *UploadSessionService) SlaveDelete(ctx context.Context, c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewAnonymousFileSystem()
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	session, ok := cache.Get(filesystem.UploadSessionCachePrefix + service.ID)
	if !ok {
		return serializer.Err(serializer.CodeUploadSessionExpired, "", nil)
	}

	if _, err := fs.Handler.Delete(ctx, []string{session.(serializer.UploadSession).SavePath}); err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "Failed to delete temp file", err)
	}

	cache.Deletes([]string{service.ID}, filesystem.UploadSessionCachePrefix)
	return serializer.Response{}
}

// DeleteAllUploadSession 删除当前用户的全部上传绘会话
func DeleteAllUploadSession(ctx context.Context, c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	// 查找需要删除的上传会话的占位文件
	files := model.GetUploadPlaceholderFiles(fs.User.ID)
	fileIDs := make([]uint, len(files))
	for i, file := range files {
		fileIDs[i] = file.ID
	}

	// 删除文件
	if err := fs.Delete(ctx, []uint{}, fileIDs, false, false); err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "Failed to cleanup upload session", err)
	}

	return serializer.Response{}
}
