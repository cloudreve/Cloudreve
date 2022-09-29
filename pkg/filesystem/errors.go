package filesystem

import (
	"errors"

	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
)

var (
	ErrUnknownPolicyType        = serializer.NewError(serializer.CodeInternalSetting, "Unknown policy type", nil)
	ErrFileSizeTooBig           = serializer.NewError(serializer.CodeFileTooLarge, "", nil)
	ErrFileExtensionNotAllowed  = serializer.NewError(serializer.CodeFileTypeNotAllowed, "", nil)
	ErrInsufficientCapacity     = serializer.NewError(serializer.CodeInsufficientCapacity, "", nil)
	ErrIllegalObjectName        = serializer.NewError(serializer.CodeIllegalObjectName, "", nil)
	ErrClientCanceled           = errors.New("Client canceled operation")
	ErrRootProtected            = serializer.NewError(serializer.CodeRootProtected, "", nil)
	ErrInsertFileRecord         = serializer.NewError(serializer.CodeDBError, "Failed to create file record", nil)
	ErrFileExisted              = serializer.NewError(serializer.CodeObjectExist, "", nil)
	ErrFileUploadSessionExisted = serializer.NewError(serializer.CodeConflictUploadOngoing, "", nil)
	ErrPathNotExist             = serializer.NewError(serializer.CodeParentNotExist, "", nil)
	ErrObjectNotExist           = serializer.NewError(serializer.CodeParentNotExist, "", nil)
	ErrIO                       = serializer.NewError(serializer.CodeIOFailed, "Failed to read file data", nil)
	ErrDBListObjects            = serializer.NewError(serializer.CodeDBError, "Failed to list object records", nil)
	ErrDBDeleteObjects          = serializer.NewError(serializer.CodeDBError, "Failed to delete object records", nil)
	ErrOneObjectOnly            = serializer.ParamErr("You can only copy one object at the same time", nil)
)
