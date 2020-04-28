package response

import (
	"io"
	"time"
)

// ContentResponse 获取文件内容类方法的通用返回值。
// 有些上传策略需要重定向，
// 有些直接写文件数据到浏览器
type ContentResponse struct {
	Redirect bool
	Content  RSCloser
	URL      string
	MaxAge   int
}

// RSCloser 存储策略适配器返回的文件流，有些策略需要带有Closer
type RSCloser interface {
	io.ReadSeeker
	io.Closer
}

// Object 列出文件、目录时返回的对象
type Object struct {
	Name         string    `json:"name"`
	RelativePath string    `json:"relative_path"`
	Source       string    `json:"source"`
	Size         uint64    `json:"size"`
	IsDir        bool      `json:"is_dir"`
	LastModify   time.Time `json:"last_modify"`
}
