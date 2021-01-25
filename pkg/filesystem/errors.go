package filesystem

import (
	"errors"

	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
)

var (
	ErrUnknownPolicyType       = errors.New("未知存储策略类型")
	ErrFileSizeTooBig          = errors.New("单个文件尺寸太大")
	ErrFileExtensionNotAllowed = errors.New("不允许上传此类型的文件")
	ErrInsufficientCapacity    = errors.New("容量空间不足")
	ErrIllegalObjectName       = errors.New("目标名称非法")
	ErrClientCanceled          = errors.New("客户端取消操作")
	ErrRootProtected           = errors.New("无法对根目录进行操作")
	ErrInsertFileRecord        = serializer.NewError(serializer.CodeDBError, "无法插入文件记录", nil)
	ErrFileExisted             = serializer.NewError(serializer.CodeObjectExist, "同名文件或目录已存在", nil)
	ErrFolderExisted           = serializer.NewError(serializer.CodeObjectExist, "同名目录已存在", nil)
	ErrPathNotExist            = serializer.NewError(404, "路径不存在", nil)
	ErrObjectNotExist          = serializer.NewError(404, "文件不存在", nil)
	ErrIO                      = serializer.NewError(serializer.CodeIOFailed, "无法读取文件数据", nil)
	ErrDBListObjects           = serializer.NewError(serializer.CodeDBError, "无法列取对象记录", nil)
	ErrDBDeleteObjects         = serializer.NewError(serializer.CodeDBError, "无法删除对象记录", nil)
)
