package filesystem

import (
	"errors"
	"github.com/HFO4/cloudreve/pkg/serializer"
)

var (
	ErrUnknownPolicyType       = errors.New("未知存储策略类型")
	ErrFileSizeTooBig          = errors.New("单个文件尺寸太大")
	ErrFileExtensionNotAllowed = errors.New("不允许上传此类型的文件")
	ErrInsufficientCapacity    = errors.New("容量空间不足")
	ErrIllegalObjectName       = errors.New("目标名称非法")
	ErrInsertFileRecord        = serializer.NewError(serializer.CodeDBError, "无法插入文件记录", nil)
	ErrFileExisted             = errors.New("同名文件已存在")
	ErrPathNotExist            = serializer.NewError(404, "路径不存在", nil)
	ErrObjectNotExist          = serializer.NewError(404, "文件不存在", nil)
	ErrIO                      = serializer.NewError(serializer.CodeIOFailed, "无法读取文件数据", nil)
)
