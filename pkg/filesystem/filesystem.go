package filesystem

import (
	"context"
	"github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/auth"
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/HFO4/cloudreve/pkg/filesystem/local"
	"github.com/HFO4/cloudreve/pkg/filesystem/oss"
	"github.com/HFO4/cloudreve/pkg/filesystem/qiniu"
	"github.com/HFO4/cloudreve/pkg/filesystem/remote"
	"github.com/HFO4/cloudreve/pkg/filesystem/response"
	"github.com/HFO4/cloudreve/pkg/filesystem/upyun"
	"github.com/HFO4/cloudreve/pkg/request"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/gin-gonic/gin"
	"io"
	"net/url"
	"sync"
)

// FSPool 文件系统资源池
var FSPool = sync.Pool{
	New: func() interface{} {
		return &FileSystem{}
	},
}

// FileHeader 上传来的文件数据处理器
type FileHeader interface {
	io.Reader
	io.Closer
	GetSize() uint64
	GetMIMEType() string
	GetFileName() string
	GetVirtualPath() string
}

// Driver 存储策略适配器
type Handler interface {
	// 上传文件, dst为文件存储路径，size 为文件大小。上下文关闭
	// 时，应取消上传并清理临时文件
	Put(ctx context.Context, file io.ReadCloser, dst string, size uint64) error

	// 删除一个或多个给定路径的文件，返回删除失败的文件路径列表及错误
	Delete(ctx context.Context, files []string) ([]string, error)

	// 获取文件内容
	Get(ctx context.Context, path string) (response.RSCloser, error)

	// 获取缩略图，可直接在ContentResponse中返回文件数据流，也可指
	// 定为重定向
	Thumb(ctx context.Context, path string) (*response.ContentResponse, error)

	// 获取外链/下载地址，
	// url - 站点本身地址,
	// isDownload - 是否直接下载
	Source(ctx context.Context, path string, url url.URL, ttl int64, isDownload bool, speed int) (string, error)

	// Token 获取有效期为ttl的上传凭证和签名，同时回调会话有效期为sessionTTL
	Token(ctx context.Context, ttl int64, callbackKey string) (serializer.UploadCredential, error)
}

// FileSystem 管理文件的文件系统
type FileSystem struct {
	// 文件系统所有者
	User *model.User
	// 操作文件使用的存储策略
	Policy *model.Policy
	// 当前正在处理的文件对象
	FileTarget []model.File
	// 当前正在处理的目录对象
	DirTarget []model.Folder

	/*
	   钩子函数
	*/
	Hooks map[string][]Hook

	/*
	   文件系统处理适配器
	*/
	Handler Handler
}

// getEmptyFS 从pool中获取新的FileSystem
func getEmptyFS() *FileSystem {
	fs := FSPool.Get().(*FileSystem)
	return fs
}

// Recycle 回收FileSystem资源
func (fs *FileSystem) Recycle() {
	fs.reset()
	FSPool.Put(fs)
}

// reset 重设文件系统，以便回收使用
func (fs *FileSystem) reset() {
	fs.User = nil
	fs.CleanTargets()
	fs.Policy = nil
	fs.Hooks = nil
	fs.Handler = nil
}

// NewFileSystem 初始化一个文件系统
func NewFileSystem(user *model.User) (*FileSystem, error) {
	fs := getEmptyFS()
	fs.User = user
	// 分配存储策略适配器
	err := fs.dispatchHandler()

	// TODO 分配默认钩子
	return fs, err
}

// NewAnonymousFileSystem 初始化匿名文件系统
func NewAnonymousFileSystem() (*FileSystem, error) {
	fs := getEmptyFS()
	fs.User = &model.User{}

	// 如果是主机模式下，则为匿名文件系统分配游客用户组
	if conf.SystemConfig.Mode == "master" {
		anonymousGroup, err := model.GetGroupByID(3)
		if err != nil {
			return nil, err
		}
		fs.User.Group = anonymousGroup
	} else {
		// 从机模式下，分配本地策略处理器
		fs.Handler = local.Driver{}
	}

	return fs, nil
}

// dispatchHandler 根据存储策略分配文件适配器
// TODO 完善测试
func (fs *FileSystem) dispatchHandler() error {
	var policyType string
	var currentPolicy *model.Policy

	if fs.Policy == nil {
		// 如果没有具体指定，就是用用户当前存储策略
		policyType = fs.User.Policy.Type
		currentPolicy = &fs.User.Policy
	} else {
		policyType = fs.Policy.Type
		currentPolicy = fs.Policy
	}

	switch policyType {
	case "mock":
		return nil
	case "local":
		fs.Handler = local.Driver{
			Policy: currentPolicy,
		}
		return nil
	case "remote":
		fs.Handler = remote.Driver{
			Policy:       currentPolicy,
			Client:       request.HTTPClient{},
			AuthInstance: auth.HMACAuth{[]byte(currentPolicy.SecretKey)},
		}
		return nil
	case "qiniu":
		fs.Handler = qiniu.Driver{
			Policy: currentPolicy,
		}
		return nil
	case "oss":
		fs.Handler = oss.Driver{
			Policy:     currentPolicy,
			HTTPClient: request.HTTPClient{},
		}
		return nil
	case "upyun":
		fs.Handler = upyun.Driver{
			Policy: currentPolicy,
		}
		return nil
	default:
		return ErrUnknownPolicyType
	}
}

// NewFileSystemFromContext 从gin.Context创建文件系统
// TODO 用户不存在时使用匿名文件系统
func NewFileSystemFromContext(c *gin.Context) (*FileSystem, error) {
	user, exist := c.Get("user")
	if !exist {
		return NewAnonymousFileSystem()
	}
	fs, err := NewFileSystem(user.(*model.User))
	return fs, err
}

// SetTargetFile 设置当前处理的目标文件
func (fs *FileSystem) SetTargetFile(files *[]model.File) {
	if len(fs.FileTarget) == 0 {
		fs.FileTarget = *files
	} else {
		fs.FileTarget = append(fs.FileTarget, *files...)
	}

}

// SetTargetDir 设置当前处理的目标目录
func (fs *FileSystem) SetTargetDir(dirs *[]model.Folder) {
	if len(fs.DirTarget) == 0 {
		fs.DirTarget = *dirs
	} else {
		fs.DirTarget = append(fs.DirTarget, *dirs...)
	}

}

// SetTargetFileByIDs 根据文件ID设置目标文件，忽略用户ID
func (fs *FileSystem) SetTargetFileByIDs(ids []uint) error {
	files, err := model.GetFilesByIDs(ids, 0)
	if err != nil || len(files) == 0 {
		return ErrFileExisted.WithError(err)
	}
	fs.SetTargetFile(&files)
	return nil
}

// CleanTargets 清空目标
func (fs *FileSystem) CleanTargets() {
	fs.FileTarget = fs.FileTarget[:0]
	fs.DirTarget = fs.DirTarget[:0]
}
