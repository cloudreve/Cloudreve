package fsctx

type key int

const (
	// GinCtx Gin的上下文
	GinCtx key = iota
	// SavePathCtx 文件物理路径
	SavePathCtx
	// FileHeaderCtx 上传的文件
	FileHeaderCtx
	// PathCtx 文件或目录的虚拟路径
	PathCtx
	// FileModelCtx 文件数据库模型
	FileModelCtx
	// FolderModelCtx 目录数据库模型
	FolderModelCtx
	// HTTPCtx HTTP请求的上下文
	HTTPCtx
	// UploadPolicyCtx 上传策略，一般为slave模式下使用
	UploadPolicyCtx
	// UserCtx 用户
	UserCtx
	// ThumbSizeCtx 缩略图尺寸
	ThumbSizeCtx
	// FileSizeCtx 文件大小
	FileSizeCtx
	// ShareKeyCtx 分享文件的 HashID
	ShareKeyCtx
	// LimitParentCtx 限制父目录
	LimitParentCtx
	// IgnoreConflictCtx 忽略重名冲突
	IgnoreConflictCtx
	// RetryCtx 失败重试次数
	RetryCtx
	// ForceUsePublicEndpointCtx 强制使用公网 Endpoint
	ForceUsePublicEndpointCtx
	// CancelFuncCtx Context 取消函數
	CancelFuncCtx
	// ValidateCapacityOnceCtx 限定归还容量的操作只執行一次
	ValidateCapacityOnceCtx
)
