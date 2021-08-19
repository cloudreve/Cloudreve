package cluster

import "errors"

var (
	ErrFeatureNotExist = errors.New("No nodes in nodepool match the feature specificed")
)
