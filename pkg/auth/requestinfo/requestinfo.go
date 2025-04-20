package requestinfo

import (
	"context"
)

// RequestInfoCtx context key for RequestInfo
type RequestInfoCtx struct{}

// RequestInfoFromContext retrieves RequestInfo from context
func RequestInfoFromContext(ctx context.Context) *RequestInfo {
	v, ok := ctx.Value(RequestInfoCtx{}).(*RequestInfo)
	if !ok {
		return nil
	}

	return v
}

// RequestInfo store request info for audit
type RequestInfo struct {
	Host      string
	IP        string
	UserAgent string
}
