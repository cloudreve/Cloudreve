package filesystem

import (
	"context"
	"github.com/HFO4/cloudreve/models"
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
	BeforeUpload func(ctx context.Context, fs *FileSystem, file FileData) error
	// 上传文件后
	AfterUpload func(ctx context.Context, fs *FileSystem) error
	// 文件验证失败后
	ValidateFailed func(ctx context.Context, fs *FileSystem) error

	/*
	  文件系统处理适配器
	*/

}

// NewFileSystem 初始化一个文件系统
func NewFileSystem(user *model.User) *FileSystem {
	return &FileSystem{
		User: user,
	}
}

// Upload 上传文件
func (fs *FileSystem) Upload(ctx context.Context, file FileData) (err error) {
	err = fs.BeforeUpload(ctx, fs, file)
	if err != nil {
		return err
	}
	return nil
}
