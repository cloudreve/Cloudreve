package filesystem

import (
	"context"
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
