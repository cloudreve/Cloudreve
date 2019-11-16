package filesystem

import (
	"cloudreve/models"
	"io"
)

// FileData 上传来的文件数据处理器
type FileData interface {
	io.Reader
	io.Closer
	GetSize() uint64
	GetMIMEType() string
}

// FileSystem 管理文件的文件系统
type FileSystem struct {
	// 文件系统所有者
	User *model.User

	// 文件系统处理适配器

}

// Upload 上传文件
func (fs *FileSystem) Upload(File FileData) (err error) {
	return nil
}
