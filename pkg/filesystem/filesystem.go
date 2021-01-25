package filesystem

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sync"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/cos"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/local"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/onedrive"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/oss"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/qiniu"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/remote"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/s3"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/upyun"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/response"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/gin-gonic/gin"
	cossdk "github.com/tencentyun/cos-go-sdk-v5"
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

// Handler 存储策略适配器
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

	// List 递归列取远程端path路径下文件、目录，不包含path本身，
	// 返回的对象路径以path作为起始根目录.
	// recursive - 是否递归列出
	List(ctx context.Context, path string, recursive bool) ([]response.Object, error)
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
	// 相对根目录
	Root *model.Folder
	// 互斥锁
	Lock sync.Mutex

	/*
	   钩子函数
	*/
	Hooks map[string][]Hook

	/*
	   文件系统处理适配器
	*/
	Handler Handler

	// 回收锁
	recycleLock sync.Mutex
}

// getEmptyFS 从pool中获取新的FileSystem
func getEmptyFS() *FileSystem {
	fs := FSPool.Get().(*FileSystem)
	return fs
}

// Recycle 回收FileSystem资源
func (fs *FileSystem) Recycle() {
	fs.recycleLock.Lock()
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
	fs.Root = nil
	fs.Lock = sync.Mutex{}
	fs.recycleLock = sync.Mutex{}
}

// NewFileSystem 初始化一个文件系统
func NewFileSystem(user *model.User) (*FileSystem, error) {
	fs := getEmptyFS()
	fs.User = user
	// 分配存储策略适配器
	err := fs.DispatchHandler()

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

// DispatchHandler 根据存储策略分配文件适配器
// TODO 完善测试
func (fs *FileSystem) DispatchHandler() error {
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
	case "mock", "anonymous":
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
	case "onedrive":
		client, err := onedrive.NewClient(currentPolicy)
		fs.Handler = onedrive.Driver{
			Policy:     currentPolicy,
			Client:     client,
			HTTPClient: request.HTTPClient{},
		}
		return err
	case "cos":
		u, _ := url.Parse(currentPolicy.Server)
		b := &cossdk.BaseURL{BucketURL: u}
		fs.Handler = cos.Driver{
			Policy: currentPolicy,
			Client: cossdk.NewClient(b, &http.Client{
				Transport: &cossdk.AuthorizationTransport{
					SecretID:  currentPolicy.AccessKey,
					SecretKey: currentPolicy.SecretKey,
				},
			}),
			HTTPClient: request.HTTPClient{},
		}
		return nil
	case "s3":
		fs.Handler = s3.Driver{
			Policy: currentPolicy,
		}
		return nil
	default:
		return ErrUnknownPolicyType
	}
}

// NewFileSystemFromContext 从gin.Context创建文件系统
func NewFileSystemFromContext(c *gin.Context) (*FileSystem, error) {
	user, exist := c.Get("user")
	if !exist {
		return NewAnonymousFileSystem()
	}
	fs, err := NewFileSystem(user.(*model.User))
	return fs, err
}

// NewFileSystemFromCallback 从gin.Context创建回调用文件系统
func NewFileSystemFromCallback(c *gin.Context) (*FileSystem, error) {
	fs, err := NewFileSystemFromContext(c)
	if err != nil {
		return nil, err
	}

	// 获取回调会话
	callbackSessionRaw, ok := c.Get("callbackSession")
	if !ok {
		return nil, errors.New("找不到回调会话")
	}
	callbackSession := callbackSessionRaw.(*serializer.UploadSession)

	// 重新指向上传策略
	policy, err := model.GetPolicyByID(callbackSession.PolicyID)
	if err != nil {
		return nil, err
	}
	fs.Policy = &policy
	fs.User.Policy = policy
	err = fs.DispatchHandler()

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

// SetTargetByInterface 根据 model.File 或者 model.Folder 设置目标对象
// TODO 测试
func (fs *FileSystem) SetTargetByInterface(target interface{}) error {
	if file, ok := target.(*model.File); ok {
		fs.SetTargetFile(&[]model.File{*file})
		return nil
	}
	if folder, ok := target.(*model.Folder); ok {
		fs.SetTargetDir(&[]model.Folder{*folder})
		return nil
	}

	return ErrObjectNotExist
}

// CleanTargets 清空目标
func (fs *FileSystem) CleanTargets() {
	fs.FileTarget = fs.FileTarget[:0]
	fs.DirTarget = fs.DirTarget[:0]
}
