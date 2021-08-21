package slave

import "github.com/cloudreve/Cloudreve/v3/pkg/serializer"

var (
	ErrMasterNotFound = serializer.NewError(serializer.CodeMasterNotFound, "未知的主机节点", nil)
)
