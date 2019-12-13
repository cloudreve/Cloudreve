package response

import "io"

// ContentResponse 获取文件内容类方法的通用返回值。
// 有些上传策略需要重定向，
// 有些直接写文件数据到浏览器
type ContentResponse struct {
	Redirect bool
	Content  io.ReadSeeker
	URL      string
}

// 存储策略适配器返回的文件流，有些策略需要带有Closer
type RSCloser interface {
	io.ReadSeeker
	io.Closer
}
