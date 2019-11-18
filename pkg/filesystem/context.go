package filesystem

type key int

const (
	// GinCtx Gin的上下文
	GinCtx key = iota
	// SavePathCtx 文件物理路径
	SavePathCtx
	// FileCtx 上传的文件
	FileCtx
)
