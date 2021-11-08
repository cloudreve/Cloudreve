package cluster

import (
	"errors"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
)

var (
	ErrFeatureNotExist = errors.New("No nodes in nodepool match the feature specificed")
	ErrIlegalPath      = errors.New("path out of boundary of setting temp folder")
	ErrMasterNotFound  = serializer.NewError(serializer.CodeMasterNotFound, "未知的主机节点", nil)
)
