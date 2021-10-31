package slaveinmaster

import "errors"

var (
	ErrNotImplemented       = errors.New("this method of shadowed policy is not implemented")
	ErrSlaveSrcPathNotExist = errors.New("cannot determine source file path in slave node")
	ErrWaitResultTimeout    = errors.New("timeout waiting for slave transfer result")
)
