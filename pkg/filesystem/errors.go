package filesystem

import "errors"

var (
	ErrUnknownPolicyType       = errors.New("未知存储策略类型")
	ErrFileSizeTooBig          = errors.New("单个文件尺寸太大")
	ErrFileExtensionNotAllowed = errors.New("不允许上传此类型的文件")
	ErrInsufficientCapacity    = errors.New("容量空间不足")
	ErrIllegalObjectName       = errors.New("目标名称非法")
	ErrInsertFileRecord        = errors.New("无法插入文件记录")
	ErrFileExisted             = errors.New("同名文件已存在")
	ErrPathNotExist            = errors.New("路径不存在")
)
