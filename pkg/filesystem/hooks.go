package filesystem

import (
	"context"
	"errors"
	"io/ioutil"
	"strings"
	"sync"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

// Hook 钩子函数
type Hook func(ctx context.Context, fs *FileSystem, file fsctx.FileHeader) error

// Use 注入钩子
func (fs *FileSystem) Use(name string, hook Hook) {
	if fs.Hooks == nil {
		fs.Hooks = make(map[string][]Hook)
	}
	if _, ok := fs.Hooks[name]; ok {
		fs.Hooks[name] = append(fs.Hooks[name], hook)
		return
	}
	fs.Hooks[name] = []Hook{hook}
}

// CleanHooks 清空钩子,name为空表示全部清空
func (fs *FileSystem) CleanHooks(name string) {
	if name == "" {
		fs.Hooks = nil
	} else {
		delete(fs.Hooks, name)
	}
}

// Trigger 触发钩子,遇到第一个错误时
// 返回错误，后续钩子不会继续执行
func (fs *FileSystem) Trigger(ctx context.Context, name string, file fsctx.FileHeader) error {
	if hooks, ok := fs.Hooks[name]; ok {
		for _, hook := range hooks {
			err := hook(ctx, fs, file)
			if err != nil {
				util.Log().Warning("钩子执行失败：%s", err)
				return err
			}
		}
	}
	return nil
}

// HookSlaveUploadValidate Slave模式下对文件上传的一系列验证
func HookSlaveUploadValidate(ctx context.Context, fs *FileSystem, file fsctx.FileHeader) error {
	policy := ctx.Value(fsctx.UploadPolicyCtx).(serializer.UploadPolicy)
	fileInfo := file.Info()

	// 验证单文件尺寸
	if policy.MaxSize > 0 {
		if fileInfo.Size > policy.MaxSize {
			return ErrFileSizeTooBig
		}
	}

	// 验证文件名
	if !fs.ValidateLegalName(ctx, fileInfo.FileName) {
		return ErrIllegalObjectName
	}

	// 验证扩展名
	if len(policy.AllowedExtension) > 0 && !IsInExtensionList(policy.AllowedExtension, fileInfo.FileName) {
		return ErrFileExtensionNotAllowed
	}

	return nil
}

// HookValidateFile 一系列对文件检验的集合
func HookValidateFile(ctx context.Context, fs *FileSystem, file fsctx.FileHeader) error {
	fileInfo := file.Info()

	// 验证单文件尺寸
	if !fs.ValidateFileSize(ctx, fileInfo.Size) {
		return ErrFileSizeTooBig
	}

	// 验证文件名
	if !fs.ValidateLegalName(ctx, fileInfo.FileName) {
		return ErrIllegalObjectName
	}

	// 验证扩展名
	if !fs.ValidateExtension(ctx, fileInfo.FileName) {
		return ErrFileExtensionNotAllowed
	}

	return nil

}

// HookResetPolicy 重设存储策略为上下文已有文件
func HookResetPolicy(ctx context.Context, fs *FileSystem, file fsctx.FileHeader) error {
	originFile, ok := ctx.Value(fsctx.FileModelCtx).(model.File)
	if !ok {
		return ErrObjectNotExist
	}

	fs.Policy = originFile.GetPolicy()
	return fs.DispatchHandler()
}

// HookValidateCapacity 验证并扣除用户容量，包含数据库操作
func HookValidateCapacity(ctx context.Context, fs *FileSystem, file fsctx.FileHeader) error {
	// 验证并扣除容量
	if !fs.ValidateCapacity(ctx, file.Info().Size) {
		return ErrInsufficientCapacity
	}
	return nil
}

// HookValidateCapacityWithoutIncrease 验证用户容量，不扣除
func HookValidateCapacityWithoutIncrease(ctx context.Context, fs *FileSystem, file fsctx.FileHeader) error {
	// 验证并扣除容量
	if fs.User.GetRemainingCapacity() < file.Info().Size {
		return ErrInsufficientCapacity
	}
	return nil
}

// HookChangeCapacity 根据原有文件和新文件的大小更新用户容量
func HookChangeCapacity(ctx context.Context, fs *FileSystem, newFile fsctx.FileHeader) error {
	originFile := ctx.Value(fsctx.FileModelCtx).(model.File)
	newFileSize := newFile.Info().Size

	if newFileSize > originFile.Size {
		if !fs.ValidateCapacity(ctx, newFileSize-originFile.Size) {
			return ErrInsufficientCapacity
		}
		return nil
	}

	fs.User.DeductionStorage(originFile.Size - newFileSize)
	return nil
}

// HookDeleteTempFile 删除已保存的临时文件
func HookDeleteTempFile(ctx context.Context, fs *FileSystem, file fsctx.FileHeader) error {
	// 删除临时文件
	_, err := fs.Handler.Delete(ctx, []string{file.Info().SavePath})
	if err != nil {
		util.Log().Warning("无法清理上传临时文件，%s", err)
	}

	return nil
}

// HookCleanFileContent 清空文件内容
func HookCleanFileContent(ctx context.Context, fs *FileSystem, file fsctx.FileHeader) error {
	// 清空内容
	return fs.Handler.Put(ctx, &fsctx.FileStream{
		File:     ioutil.NopCloser(strings.NewReader("")),
		SavePath: file.Info().SavePath,
		Size:     0,
	})
}

// HookClearFileSize 将原始文件的尺寸设为0
func HookClearFileSize(ctx context.Context, fs *FileSystem, file fsctx.FileHeader) error {
	originFile, ok := ctx.Value(fsctx.FileModelCtx).(model.File)
	if !ok {
		return ErrObjectNotExist
	}
	return originFile.UpdateSize(0)
}

// HookCancelContext 取消上下文
func HookCancelContext(ctx context.Context, fs *FileSystem, file fsctx.FileHeader) error {
	cancelFunc, ok := ctx.Value(fsctx.CancelFuncCtx).(context.CancelFunc)
	if ok {
		cancelFunc()
	}
	return nil
}

// HookGiveBackCapacity 归还用户容量
func HookGiveBackCapacity(ctx context.Context, fs *FileSystem, file fsctx.FileHeader) error {
	once, ok := ctx.Value(fsctx.ValidateCapacityOnceCtx).(*sync.Once)
	if !ok {
		once = &sync.Once{}
	}

	// 归还用户容量
	res := true
	once.Do(func() {
		res = fs.User.DeductionStorage(file.Info().Size)
	})

	if !res {
		return errors.New("无法继续降低用户已用存储")
	}
	return nil
}

// HookUpdateSourceName 更新文件SourceName
// TODO：测试
func HookUpdateSourceName(ctx context.Context, fs *FileSystem, file fsctx.FileHeader) error {
	originFile, ok := ctx.Value(fsctx.FileModelCtx).(model.File)
	if !ok {
		return ErrObjectNotExist
	}
	return originFile.UpdateSourceName(originFile.SourceName)
}

// GenericAfterUpdate 文件内容更新后
func GenericAfterUpdate(ctx context.Context, fs *FileSystem, newFile fsctx.FileHeader) error {
	// 更新文件尺寸
	originFile, ok := ctx.Value(fsctx.FileModelCtx).(model.File)
	if !ok {
		return ErrObjectNotExist
	}

	fs.SetTargetFile(&[]model.File{originFile})

	err := originFile.UpdateSize(newFile.Info().Size)
	if err != nil {
		return err
	}

	// 尝试清空原有缩略图并重新生成
	if originFile.GetPolicy().IsThumbGenerateNeeded() {
		fs.recycleLock.Lock()
		go func() {
			defer fs.recycleLock.Unlock()
			if originFile.PicInfo != "" {
				_, _ = fs.Handler.Delete(ctx, []string{originFile.SourceName + conf.ThumbConfig.FileSuffix})
				fs.GenerateThumbnail(ctx, &originFile)
			}
		}()
	}

	return nil
}

// SlaveAfterUpload Slave模式下上传完成钩子
func SlaveAfterUpload(ctx context.Context, fs *FileSystem, fileHeader fsctx.FileHeader) error {
	policy := ctx.Value(fsctx.UploadPolicyCtx).(serializer.UploadPolicy)
	fileInfo := fileHeader.Info()

	// 构造一个model.File，用于生成缩略图
	file := model.File{
		Name:       fileInfo.FileName,
		SourceName: fileInfo.SavePath,
	}
	fs.GenerateThumbnail(ctx, &file)

	if policy.CallbackURL == "" {
		return nil
	}

	// 发送回调请求
	callbackBody := serializer.UploadCallback{
		Name:       file.Name,
		SourceName: file.SourceName,
		PicInfo:    file.PicInfo,
		Size:       fileInfo.Size,
	}
	return request.RemoteCallback(policy.CallbackURL, callbackBody)
}

// GenericAfterUpload 文件上传完成后，包含数据库操作
func GenericAfterUpload(ctx context.Context, fs *FileSystem, fileHeader fsctx.FileHeader) error {
	fileInfo := fileHeader.Info()

	// 检查路径是否存在，不存在就创建
	isExist, folder := fs.IsPathExist(fileInfo.VirtualPath)
	if !isExist {
		newFolder, err := fs.CreateDirectory(ctx, fileInfo.VirtualPath)
		if err != nil {
			return err
		}
		folder = newFolder
	}

	// 检查文件是否存在
	if ok, file := fs.IsChildFileExist(
		folder,
		fileInfo.FileName,
	); ok {
		if file.UploadSessionID != nil {
			return ErrFileUploadSessionExisted
		}

		return ErrFileExisted
	}

	// 向数据库中插入记录
	file, err := fs.AddFile(ctx, folder, fileHeader)
	if err != nil {
		return ErrInsertFileRecord
	}
	fs.SetTargetFile(&[]model.File{*file})

	// 异步尝试生成缩略图
	if fs.Policy.IsThumbGenerateNeeded() {
		fs.recycleLock.Lock()
		go func() {
			defer fs.recycleLock.Unlock()
			fs.GenerateThumbnail(ctx, file)
		}()
	}

	return nil
}

// HookClearFileHeaderSize 将FileHeader大小设定为0
func HookClearFileHeaderSize(ctx context.Context, fs *FileSystem, fileHeader fsctx.FileHeader) error {
	fileHeader.SetSize(0)
	return nil
}
