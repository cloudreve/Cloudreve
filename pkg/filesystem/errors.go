package filesystem

import (
	"errors"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
)

var (
	ErrUnknownPolicyType        = serializer.NewError(serializer.CodeInternalSetting, "Unknown policy type", nil)
	ErrFileSizeTooBig           = serializer.NewError(serializer.CodeFileTooLarge, "File is too large", nil)
	ErrFileExtensionNotAllowed  = serializer.NewError(serializer.CodeFileTypeNotAllowed, "File type not allowed", nil)
	ErrInsufficientCapacity     = serializer.NewError(serializer.CodeInsufficientCapacity, "Insufficient capacity", nil)
	ErrIllegalObjectName        = serializer.NewError(serializer.CodeIllegalObjectName, "Invalid object name", nil)
	ErrClientCanceled           = errors.New("Client canceled operation")
	ErrRootProtected            = serializer.NewError(serializer.CodeRootProtected, "Root protected", nil)
	ErrInsertFileRecord         = serializer.NewError(serializer.CodeDBError, "Failed to create file record", nil)
	ErrFileExisted              = serializer.NewError(serializer.CodeObjectExist, "Object existed", nil)
	ErrFileUploadSessionExisted = serializer.NewError(serializer.CodeConflictUploadOngoing, "Upload session existed", nil)
	ErrPathNotExist             = serializer.NewError(serializer.CodeParentNotExist, "Path not exist", nil)
	ErrObjectNotExist           = serializer.NewError(serializer.CodeParentNotExist, "Object not exist", nil)
	ErrIO                       = serializer.NewError(serializer.CodeIOFailed, "Failed to read file data", nil)
	ErrDBListObjects            = serializer.NewError(serializer.CodeDBError, "Failed to list object records", nil)
	ErrDBDeleteObjects          = serializer.NewError(serializer.CodeDBError, "Failed to delete object records", nil)
	ErrOneObjectOnly            = serializer.ParamErr("You can only copy one object at the same time", nil)
)
