package filesystem

import (
	"context"
	"errors"
)

// GenericBeforeUpload 通用上传前处理钩子，包含数据库操作
func GenericBeforeUpload(ctx context.Context, fs *FileSystem, file FileData) error {
	// 验证单文件尺寸
	if !fs.ValidateFileSize(ctx, file.GetSize()) {
		return errors.New("单个文件尺寸太大")
	}

	// 验证并扣除容量
	if !fs.ValidateCapacity(ctx, file.GetSize()) {
		return errors.New("容量空间不足")
	}

	// 验证扩展名
	if !fs.ValidateExtension(ctx, file.GetFileName()) {
		return errors.New("不允许上传此类型的文件")
	}
	return nil
}
