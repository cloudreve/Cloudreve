package filesystem

import (
	"context"
	"os"
	"path"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

/* ================
	 上传处理相关
   ================
*/

const (
	UploadSessionMetaKey     = "upload_session"
	UploadSessionCtx         = "uploadSession"
	UserCtx                  = "user"
	UploadSessionCachePrefix = "callback_"
)

// Upload 上传文件
func (fs *FileSystem) Upload(ctx context.Context, file *fsctx.FileStream) (err error) {
	// 上传前的钩子
	err = fs.Trigger(ctx, "BeforeUpload", file)
	if err != nil {
		request.BlackHole(file)
		return err
	}

	// 生成文件名和路径,
	var savePath string
	if file.SavePath == "" {
		// 如果是更新操作就从上下文中获取
		if originFile, ok := ctx.Value(fsctx.FileModelCtx).(model.File); ok {
			savePath = originFile.SourceName
		} else {
			savePath = fs.GenerateSavePath(ctx, file)
		}
		file.SavePath = savePath
	}

	// 保存文件
	if file.Mode&fsctx.Nop != fsctx.Nop {
		// 处理客户端未完成上传时，关闭连接
		go fs.CancelUpload(ctx, savePath, file)

		err = fs.Handler.Put(ctx, file)
		if err != nil {
			fs.Trigger(ctx, "AfterUploadFailed", file)
			return err
		}
	}

	// 上传完成后的钩子
	err = fs.Trigger(ctx, "AfterUpload", file)

	if err != nil {
		// 上传完成后续处理失败
		followUpErr := fs.Trigger(ctx, "AfterValidateFailed", file)
		// 失败后再失败...
		if followUpErr != nil {
			util.Log().Debug("AfterValidateFailed hook execution failed: %s", followUpErr)
		}

		return err
	}

	return nil
}

// GenerateSavePath 生成要存放文件的路径
// TODO 完善测试
func (fs *FileSystem) GenerateSavePath(ctx context.Context, file fsctx.FileHeader) string {
	fileInfo := file.Info()
	return path.Join(
		fs.Policy.GeneratePath(
			fs.User.Model.ID,
			fileInfo.VirtualPath,
		),
		fs.Policy.GenerateFileName(
			fs.User.Model.ID,
			fileInfo.FileName,
		),
	)

}

// CancelUpload 监测客户端取消上传
func (fs *FileSystem) CancelUpload(ctx context.Context, path string, file fsctx.FileHeader) {
	var reqContext context.Context
	if ginCtx, ok := ctx.Value(fsctx.GinCtx).(*gin.Context); ok {
		reqContext = ginCtx.Request.Context()
	} else if reqCtx, ok := ctx.Value(fsctx.HTTPCtx).(context.Context); ok {
		reqContext = reqCtx
	} else {
		return
	}

	select {
	case <-reqContext.Done():
		select {
		case <-ctx.Done():
			// 客户端正常关闭，不执行操作
		default:
			// 客户端取消上传，删除临时文件
			util.Log().Debug("Client canceled upload.")
			if fs.Hooks["AfterUploadCanceled"] == nil {
				return
			}
			err := fs.Trigger(ctx, "AfterUploadCanceled", file)
			if err != nil {
				util.Log().Debug("AfterUploadCanceled hook execution failed: %s", err)
			}
		}

	}
}

// CreateUploadSession 创建上传会话
func (fs *FileSystem) CreateUploadSession(ctx context.Context, file *fsctx.FileStream) (*serializer.UploadCredential, error) {
	// 获取相关有效期设置
	callBackSessionTTL := model.GetIntSetting("upload_session_timeout", 86400)

	callbackKey := uuid.Must(uuid.NewV4()).String()
	fileSize := file.Size

	// 创建占位的文件，同时校验文件信息
	file.Mode = fsctx.Nop
	if callbackKey != "" {
		file.UploadSessionID = &callbackKey
	}

	fs.Use("BeforeUpload", HookValidateFile)
	fs.Use("BeforeUpload", HookValidateCapacity)

	// 验证文件规格
	if err := fs.Upload(ctx, file); err != nil {
		return nil, err
	}

	uploadSession := &serializer.UploadSession{
		Key:            callbackKey,
		UID:            fs.User.ID,
		Policy:         *fs.Policy,
		VirtualPath:    file.VirtualPath,
		Name:           file.Name,
		Size:           fileSize,
		SavePath:       file.SavePath,
		LastModified:   file.LastModified,
		CallbackSecret: util.RandStringRunes(32),
	}

	// 获取上传凭证
	credential, err := fs.Handler.Token(ctx, int64(callBackSessionTTL), uploadSession, file)
	if err != nil {
		return nil, err
	}

	// 创建占位符
	if !fs.Policy.IsUploadPlaceholderWithSize() {
		fs.Use("AfterUpload", HookClearFileHeaderSize)
	}
	fs.Use("AfterUpload", GenericAfterUpload)
	ctx = context.WithValue(ctx, fsctx.IgnoreDirectoryConflictCtx, true)
	if err := fs.Upload(ctx, file); err != nil {
		return nil, err
	}

	// 创建回调会话
	err = cache.Set(
		UploadSessionCachePrefix+callbackKey,
		*uploadSession,
		callBackSessionTTL,
	)
	if err != nil {
		return nil, err
	}

	// 补全上传凭证其他信息
	credential.Expires = time.Now().Add(time.Duration(callBackSessionTTL) * time.Second).Unix()

	return credential, nil
}

// UploadFromStream 从文件流上传文件
func (fs *FileSystem) UploadFromStream(ctx context.Context, file *fsctx.FileStream, resetPolicy bool) error {
	// 给文件系统分配钩子
	fs.Lock.Lock()
	if resetPolicy {
		err := fs.SetPolicyFromPath(file.VirtualPath)
		if err != nil {
			return err
		}
	}

	if fs.Hooks == nil {
		fs.Use("BeforeUpload", HookValidateFile)
		fs.Use("BeforeUpload", HookValidateCapacity)
		fs.Use("AfterUploadCanceled", HookDeleteTempFile)
		fs.Use("AfterUpload", GenericAfterUpload)
		fs.Use("AfterValidateFailed", HookDeleteTempFile)
	}
	fs.Lock.Unlock()

	// 开始上传
	return fs.Upload(ctx, file)
}

// UploadFromPath 将本机已有文件上传到用户的文件系统
func (fs *FileSystem) UploadFromPath(ctx context.Context, src, dst string, mode fsctx.WriteMode) error {
	file, err := os.Open(util.RelativePath(src))
	if err != nil {
		return err
	}
	defer file.Close()

	// 获取源文件大小
	fi, err := file.Stat()
	if err != nil {
		return err
	}
	size := fi.Size()

	// 开始上传
	return fs.UploadFromStream(ctx, &fsctx.FileStream{
		File:        file,
		Seeker:      file,
		Size:        uint64(size),
		Name:        path.Base(dst),
		VirtualPath: path.Dir(dst),
		Mode:        mode,
	}, true)
}
