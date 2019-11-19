package filesystem

import (
	"context"
	"github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem/local"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/gin-gonic/gin"
	"io"
	"path"
	"path/filepath"
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
	Put(ctx context.Context, file io.ReadCloser, dst string) error
	// 删除一个或多个文件
	Delete(ctx context.Context, files []string) ([]string, error)
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
	BeforeUpload func(ctx context.Context, fs *FileSystem) error
	// 上传文件后
	AfterUpload func(ctx context.Context, fs *FileSystem) error
	// 文件保存成功，插入数据库验证失败后
	AfterValidateFailed func(ctx context.Context, fs *FileSystem) error
	// 用户取消上传后
	AfterUploadCanceled func(ctx context.Context, fs *FileSystem) error

	/*
	  文件系统处理适配器
	*/
	Handler Handler
}

// NewFileSystem 初始化一个文件系统
func NewFileSystem(user *model.User) (*FileSystem, error) {
	var handler Handler

	// 根据存储策略类型分配适配器
	switch user.Policy.Type {
	case "local":
		handler = local.Handler{}
	default:
		return nil, UnknownPolicyTypeError
	}

	// TODO 分配默认钩子
	return &FileSystem{
		User:    user,
		Handler: handler,
	}, nil
}

/* ============
	 文件相关
   ============
*/

// AddFile 新增文件记录
func (fs *FileSystem) AddFile(parent *model.Folder) error {
	return nil
}

/* ================
	 上传处理相关
   ================
*/

// Upload 上传文件
func (fs *FileSystem) Upload(ctx context.Context, file FileHeader) (err error) {
	ctx = context.WithValue(ctx, FileHeaderCtx, file)

	// 上传前的钩子
	if fs.BeforeUpload != nil {
		err = fs.BeforeUpload(ctx, fs)
		if err != nil {
			return err
		}
	}

	// 生成文件名和路径
	savePath := fs.GenerateSavePath(ctx, file)

	// 处理客户端未完成上传时，关闭连接
	go fs.CancelUpload(ctx, savePath, file)

	// 保存文件
	err = fs.Handler.Put(ctx, file, savePath)
	if err != nil {
		return err
	}

	// 上传完成后的钩子
	if fs.AfterUpload != nil {
		ctx = context.WithValue(ctx, SavePathCtx, savePath)
		err = fs.AfterUpload(ctx, fs)

		if err != nil {
			// 上传完成后续处理失败
			if fs.AfterValidateFailed != nil {
				followUpErr := fs.AfterValidateFailed(ctx, fs)
				// 失败后再失败...
				if followUpErr != nil {
					util.Log().Warning("AfterValidateFailed 钩子执行失败，%s", followUpErr)
				}
			}
			return err
		}
	}

	return nil
}

// GenerateSavePath 生成要存放文件的路径
func (fs *FileSystem) GenerateSavePath(ctx context.Context, file FileHeader) string {
	return filepath.Join(
		fs.User.Policy.GeneratePath(
			fs.User.Model.ID,
			file.GetVirtualPath(),
		),
		fs.User.Policy.GenerateFileName(
			fs.User.Model.ID,
			file.GetFileName(),
		),
	)
}

// CancelUpload 监测客户端取消上传
func (fs *FileSystem) CancelUpload(ctx context.Context, path string, file FileHeader) {
	ginCtx := ctx.Value(GinCtx).(*gin.Context)
	select {
	case <-ctx.Done():
		// 客户端正常关闭，不执行操作
	case <-ginCtx.Request.Context().Done():
		// 客户端取消了上传
		if fs.AfterUploadCanceled == nil {
			return
		}
		ctx = context.WithValue(ctx, SavePathCtx, path)
		err := fs.AfterUploadCanceled(ctx, fs)
		if err != nil {
			util.Log().Warning("执行 AfterUploadCanceled 钩子出错，%s", err)
		}
	}
}

/* =================
	 路径/目录相关
   =================
*/

// IsPathExist 返回给定目录是否存在
// 如果存在就返回目录
func (fs *FileSystem) IsPathExist(path string) (bool, model.Folder) {
	folder, err := model.GetFolderByPath(path, fs.User.ID)
	return err == nil, folder
}

// IsFileExist 返回给定路径的文件是否存在
func (fs *FileSystem) IsFileExist(fullPath string) bool {
	basePath := path.Dir(fullPath)
	fileName := path.Base(fullPath)

	_, err := model.GetFileByPathAndName(basePath, fileName, fs.User.ID)

	return err == nil
}
