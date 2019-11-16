package filesystem

import "errors"

// GenericBeforeUpload 通用上传前处理钩子，包含数据库操作
func GenericBeforeUpload(fs *FileSystem, file FileData) error {
	// 验证单文件尺寸
	if !fs.ValidateFileSize(file.GetSize()) {
		return errors.New("单个文件尺寸太大")
	}

	// 验证并扣除容量
	if !fs.ValidateCapacity(file.GetSize()) {
		return errors.New("容量空间不足")
	}

	// 验证扩展名
	if !fs.ValidateExtension(file.GetFileName()) {
		return errors.New("不允许上传此类型的文件")
	}
	return nil
}
