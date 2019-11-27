package filesystem

import (
	"context"
	"errors"
	"github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem/local"
	"github.com/gin-gonic/gin"
	"io"
)

// FileHeader 上传来的文件数据处理器
type FileHeader interface {
	io.Reader
	io.Closer
	GetSize() uint64
	GetMIMEType() string
	GetFileName() string
	GetVirtualPath() string
}

// Handler 存储策略适配器
type Handler interface {
	// 上传文件
	Put(ctx context.Context, file io.ReadCloser, dst string, size uint64) error
	// 删除一个或多个文件
	Delete(ctx context.Context, files []string) ([]string, error)
	// 获取文件
	Get(ctx context.Context, path string) (io.ReadSeeker, error)
}

// FileSystem 管理文件的文件系统
type FileSystem struct {
	// 文件系统所有者
	User *model.User
	// 操作文件使用的上传策略
	Policy *model.Policy
	// 当前正在处理的文件对象
	Target *model.File

	/*
	   钩子函数
	*/
	// 上传文件前
	BeforeUpload []Hook
	// 上传文件后
	AfterUpload []Hook
	// 文件保存成功，插入数据库验证失败后
	AfterValidateFailed []Hook
	// 用户取消上传后
	AfterUploadCanceled []Hook
	// 文件下载前
	BeforeFileDownload []Hook

	/*
	   文件系统处理适配器
	*/
	Handler Handler
}

// NewFileSystem 初始化一个文件系统
func NewFileSystem(user *model.User) (*FileSystem, error) {
	fs := &FileSystem{
		User: user,
	}

	// 分配存储策略适配器
	err := fs.dispatchHandler()

	// TODO 分配默认钩子
	return fs, err
}

// dispatchHandler 根据存储策略分配文件适配器
// TODO: 测试
func (fs *FileSystem) dispatchHandler() error {
	var policyType string
	if fs.Policy == nil {
		// 如果没有具体指定，就是用用户当前存储策略
		policyType = fs.User.Policy.Type
	} else {
		policyType = fs.Policy.Type
	}

	// 根据存储策略类型分配适配器
	switch policyType {
	case "local":
		fs.Handler = local.Handler{}
		return nil
	default:
		return ErrUnknownPolicyType
	}
}

// NewFileSystemFromContext 从gin.Context创建文件系统
func NewFileSystemFromContext(c *gin.Context) (*FileSystem, error) {
	user, exist := c.Get("user")
	if !exist {
		return nil, errors.New("无法找到用户")
	}
	fs, err := NewFileSystem(user.(*model.User))
	return fs, err
}
