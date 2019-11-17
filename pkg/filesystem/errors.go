package filesystem

import "errors"

var (
	UnknownPolicyTypeError       = errors.New("未知存储策略类型")
	FileSizeTooBigError          = errors.New("单个文件尺寸太大")
	FileExtensionNotAllowedError = errors.New("不允许上传此类型的文件")
	InsufficientCapacityError    = errors.New("容量空间不足")
)
