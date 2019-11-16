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
	GetFileName() string
}

// FileSystem 管理文件的文件系统
type FileSystem struct {
	/*
	  文件系统所有者
	*/
	User *model.User

	/*
	  钩子函数
	*/
	// 上传文件前
	BeforeUpload func(fs *FileSystem, file FileData) error
	// 上传文件后
	AfterUpload func(fs *FileSystem) error
	// 文件验证失败后
	ValidateFailed func(fs *FileSystem) error

	/*
	  文件系统处理适配器
	*/

}

// Upload 上传文件
func (fs *FileSystem) Upload(file FileData) (err error) {
	err = fs.BeforeUpload(fs, file)
	if err != nil {
		return err
	}
	return nil
}
