package filesystem

import (
	"context"
	"errors"
	"github.com/HFO4/cloudreve/pkg/util"
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
	filePath := ctx.Value(PathCtx).(string)
	if ok, _ := fs.IsFileExist(filePath); ok {
		return nil
	}
	return ErrObjectNotExist
}

// HookValidateFile 一系列对文件检验的集合
func HookValidateFile(ctx context.Context, fs *FileSystem) error {
	file := ctx.Value(FileHeaderCtx).(FileHeader)

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

// HookValidateCapacity 验证并扣除用户容量，包含数据库操作
func HookValidateCapacity(ctx context.Context, fs *FileSystem) error {
	file := ctx.Value(FileHeaderCtx).(FileHeader)
	// 验证并扣除容量
	if !fs.ValidateCapacity(ctx, file.GetSize()) {
		return ErrInsufficientCapacity
	}
	return nil
}

// HookDeleteTempFile 删除已保存的临时文件
func HookDeleteTempFile(ctx context.Context, fs *FileSystem) error {
	filePath := ctx.Value(SavePathCtx).(string)
	// 删除临时文件
	if util.Exists(filePath) {
		_, err := fs.Handler.Delete(ctx, []string{filePath})
		if err != nil {
			return err
		}
	}

	return nil
}

// HookGiveBackCapacity 归还用户容量
func HookGiveBackCapacity(ctx context.Context, fs *FileSystem) error {
	file := ctx.Value(FileHeaderCtx).(FileHeader)

	// 归还用户容量
	if !fs.User.DeductionStorage(file.GetSize()) {
		return errors.New("无法继续降低用户已用存储")
	}
	return nil
}

// GenericAfterUpload 文件上传完成后，包含数据库操作
func GenericAfterUpload(ctx context.Context, fs *FileSystem) error {
	// 文件存放的虚拟路径
	virtualPath := ctx.Value(FileHeaderCtx).(FileHeader).GetVirtualPath()

	// 检查路径是否存在
	isExist, folder := fs.IsPathExist(virtualPath)
	if !isExist {
		return ErrPathNotExist
	}

	// 检查文件是否存在
	if ok, _ := fs.IsChildFileExist(
		folder,
		ctx.Value(FileHeaderCtx).(FileHeader).GetFileName(),
	); ok {
		return ErrFileExisted
	}

	// 向数据库中插入记录
	file, err := fs.AddFile(ctx, folder)
	if err != nil {
		return ErrInsertFileRecord
	}

	// 异步尝试生成缩略图
	go fs.GenerateThumbnail(ctx, file)

	return nil
}
