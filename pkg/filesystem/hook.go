package filesystem

import (
	"context"
	"errors"
	"github.com/HFO4/cloudreve/pkg/util"
)

// GenericBeforeUpload 通用上传前处理钩子，包含数据库操作
func GenericBeforeUpload(ctx context.Context, fs *FileSystem, file FileData) error {
	// 验证单文件尺寸
	if !fs.ValidateFileSize(ctx, file.GetSize()) {
		return FileSizeTooBigError
	}

	// 验证文件名
	if !fs.ValidateLegalName(ctx, file.GetFileName()) {
		return IlegalObjectNameError
	}

	// 验证扩展名
	if !fs.ValidateExtension(ctx, file.GetFileName()) {
		return FileExtensionNotAllowedError
	}

	// 验证并扣除容量
	if !fs.ValidateCapacity(ctx, file.GetSize()) {
		return InsufficientCapacityError
	}
	return nil
}

// GenericAfterUploadCanceled 通用上传取消处理钩子，包含数据库操作
func GenericAfterUploadCanceled(ctx context.Context, fs *FileSystem, file FileData) error {
	filePath := ctx.Value("path").(string)
	// 删除临时文件
	if util.Exists(filePath) {
		_, err := fs.Handler.Delete(ctx, []string{filePath})
		if err != nil {
			return err
		}
	}

	// 归还用户容量
	if !fs.User.DeductionStorage(file.GetSize()) {
		return errors.New("无法继续降低用户已用存储")
	}
	return nil
}
