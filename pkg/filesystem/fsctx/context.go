package fsctx

type key int

const (
	// GinCtx Gin的上下文
	GinCtx key = iota
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
	// IgnoreDirectoryConflictCtx 忽略目录重名冲突
	IgnoreDirectoryConflictCtx
	// RetryCtx 失败重试次数
	RetryCtx
	// ForceUsePublicEndpointCtx 强制使用公网 Endpoint
	ForceUsePublicEndpointCtx
	// CancelFuncCtx Context 取消函數
	CancelFuncCtx
	// 文件在从机节点中的路径
	SlaveSrcPath
	// Webdav目标名称
	WebdavDstName
	// WebDAVCtx WebDAV
	WebDAVCtx
	// WebDAV反代Url
	WebDAVProxyUrlCtx
)
