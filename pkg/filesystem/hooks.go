package filesystem

import (
	"context"
	"errors"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/pkg/util"
	"io/ioutil"
	"strings"
)

// Hook 钩子函数
type Hook func(ctx context.Context, fs *FileSystem) error

// Use 注入钩子
func (fs *FileSystem) Use(name string, hook Hook) {
	switch name {
	case "BeforeUpload":
		fs.BeforeUpload = append(fs.BeforeUpload, hook)
	case "AfterUpload":
		fs.AfterUpload = append(fs.AfterUpload, hook)
	case "AfterValidateFailed":
		fs.AfterValidateFailed = append(fs.AfterValidateFailed, hook)
	case "AfterUploadCanceled":
		fs.AfterUploadCanceled = append(fs.AfterUploadCanceled, hook)
	case "BeforeFileDownload":
		fs.BeforeFileDownload = append(fs.BeforeFileDownload, hook)
	}
}

// Trigger 触发钩子,遇到第一个错误时
// 返回错误，后续钩子不会继续执行
func (fs *FileSystem) Trigger(ctx context.Context, hooks []Hook) error {
	for _, hook := range hooks {
		err := hook(ctx, fs)
		if err != nil {
			util.Log().Warning("钩子执行失败：%s", err)
			return err
		}
	}
	return nil
}

// HookIsFileExist 检查虚拟路径文件是否存在
func HookIsFileExist(ctx context.Context, fs *FileSystem) error {
	filePath := ctx.Value(fsctx.PathCtx).(string)
	if ok, _ := fs.IsFileExist(filePath); ok {
		return nil
	}
	return ErrObjectNotExist
}

// HookSlaveUploadValidate Slave模式下对文件上传的一系列验证
func HookSlaveUploadValidate(ctx context.Context, fs *FileSystem) error {
	file := ctx.Value(fsctx.FileHeaderCtx).(FileHeader)
	policy := ctx.Value(fsctx.UploadPolicyCtx).(serializer.UploadPolicy)

	// 验证单文件尺寸
	if file.GetSize() > policy.MaxSize {
		return ErrFileSizeTooBig
	}

	// 验证文件名
	if !fs.ValidateLegalName(ctx, file.GetFileName()) {
		return ErrIllegalObjectName
	}

	// 验证扩展名
	if len(policy.AllowedExtension) > 0 && !IsInExtensionList(policy.AllowedExtension, file.GetFileName()) {
		return ErrFileExtensionNotAllowed
	}

	return nil
}

// HookValidateFile 一系列对文件检验的集合
func HookValidateFile(ctx context.Context, fs *FileSystem) error {
	file := ctx.Value(fsctx.FileHeaderCtx).(FileHeader)

	// 验证单文件尺寸
	if !fs.ValidateFileSize(ctx, file.GetSize()) {
		return ErrFileSizeTooBig
	}

	// 验证文件名
	if !fs.ValidateLegalName(ctx, file.GetFileName()) {
		return ErrIllegalObjectName
	}

	// 验证扩展名
	if !fs.ValidateExtension(ctx, file.GetFileName()) {
		return ErrFileExtensionNotAllowed
	}

	return nil

}

// HookResetPolicy 重设存储策略为上下文已有文件
func HookResetPolicy(ctx context.Context, fs *FileSystem) error {
	originFile, ok := ctx.Value(fsctx.FileModelCtx).(model.File)
	if !ok {
		return ErrObjectNotExist
	}

	fs.Policy = originFile.GetPolicy()
	return fs.dispatchHandler()
}

// HookValidateCapacity 验证并扣除用户容量，包含数据库操作
func HookValidateCapacity(ctx context.Context, fs *FileSystem) error {
	file := ctx.Value(fsctx.FileHeaderCtx).(FileHeader)
	// 验证并扣除容量
	if !fs.ValidateCapacity(ctx, file.GetSize()) {
		return ErrInsufficientCapacity
	}
	return nil
}

// HookChangeCapacity 根据原有文件和新文件的大小更新用户容量
func HookChangeCapacity(ctx context.Context, fs *FileSystem) error {
	newFile := ctx.Value(fsctx.FileHeaderCtx).(FileHeader)
	originFile := ctx.Value(fsctx.FileModelCtx).(model.File)

	if newFile.GetSize() > originFile.Size {
		if !fs.ValidateCapacity(ctx, newFile.GetSize()-originFile.Size) {
			return ErrInsufficientCapacity
		}
		return nil
	}

	fs.User.DeductionStorage(originFile.Size - newFile.GetSize())
	return nil
}

// HookDeleteTempFile 删除已保存的临时文件
func HookDeleteTempFile(ctx context.Context, fs *FileSystem) error {
	filePath := ctx.Value(fsctx.SavePathCtx).(string)
	// 删除临时文件
	// TODO 其他策略。Exists？
	if util.Exists(filePath) {
		_, err := fs.Handler.Delete(ctx, []string{filePath})
		if err != nil {
			return err
		}
	}

	return nil
}

// HookCleanFileContent 清空文件内容
func HookCleanFileContent(ctx context.Context, fs *FileSystem) error {
	filePath := ctx.Value(fsctx.SavePathCtx).(string)
	// 清空内容
	return fs.Handler.Put(ctx, ioutil.NopCloser(strings.NewReader("")), filePath, 0)
}

// HookClearFileSize 将原始文件的尺寸设为0
func HookClearFileSize(ctx context.Context, fs *FileSystem) error {
	originFile, ok := ctx.Value(fsctx.FileModelCtx).(model.File)
	if !ok {
		return ErrObjectNotExist
	}
	return originFile.UpdateSize(0)
}

// HookGiveBackCapacity 归还用户容量
func HookGiveBackCapacity(ctx context.Context, fs *FileSystem) error {
	file := ctx.Value(fsctx.FileHeaderCtx).(FileHeader)

	// 归还用户容量
	if !fs.User.DeductionStorage(file.GetSize()) {
		return errors.New("无法继续降低用户已用存储")
	}
	return nil
}

// GenericAfterUpdate 文件内容更新后
func GenericAfterUpdate(ctx context.Context, fs *FileSystem) error {
	// 更新文件尺寸
	originFile, ok := ctx.Value(fsctx.FileModelCtx).(model.File)
	if !ok {
		return ErrObjectNotExist
	}

	fs.SetTargetFile(&[]model.File{originFile})

	newFile, ok := ctx.Value(fsctx.FileHeaderCtx).(FileHeader)
	if !ok {
		return ErrObjectNotExist
	}
	err := originFile.UpdateSize(newFile.GetSize())
	if err != nil {
		return err
	}

	// 尝试清空原有缩略图并重新生成
	go func() {
		if originFile.PicInfo != "" {
			_, _ = fs.Handler.Delete(ctx, []string{originFile.SourceName + conf.ThumbConfig.FileSuffix})
			fs.GenerateThumbnail(ctx, &originFile)
		}
	}()

	return nil
}

// SlaveAfterUpload Slave模式下上传完成钩子
// TODO 测试
func SlaveAfterUpload(ctx context.Context, fs *FileSystem) error {
	fileHeader := ctx.Value(fsctx.FileHeaderCtx).(FileHeader)
	// 构造一个model.File，用于生成缩略图
	file := model.File{
		Name:       fileHeader.GetFileName(),
		SourceName: ctx.Value(fsctx.SavePathCtx).(string),
	}
	fs.GenerateThumbnail(ctx, &file)

	// TODO 发送回调请求

	return nil
}

// GenericAfterUpload 文件上传完成后，包含数据库操作
func GenericAfterUpload(ctx context.Context, fs *FileSystem) error {
	// 文件存放的虚拟路径
	virtualPath := ctx.Value(fsctx.FileHeaderCtx).(FileHeader).GetVirtualPath()

	// 检查路径是否存在
	isExist, folder := fs.IsPathExist(virtualPath)
	if !isExist {
		return ErrPathNotExist
	}

	// 检查文件是否存在
	if ok, _ := fs.IsChildFileExist(
		folder,
		ctx.Value(fsctx.FileHeaderCtx).(FileHeader).GetFileName(),
	); ok {
		return ErrFileExisted
	}

	// 向数据库中插入记录
	file, err := fs.AddFile(ctx, folder)
	if err != nil {
		return ErrInsertFileRecord
	}
	fs.SetTargetFile(&[]model.File{*file})

	// 异步尝试生成缩略图
	go fs.GenerateThumbnail(ctx, file)

	return nil
}
