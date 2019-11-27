package filesystem

import (
	"context"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/juju/ratelimit"
	"io"
)

/* ============
	 文件相关
   ============
*/

// 限速后的ReaderSeeker
type lrs struct {
	io.ReadSeeker
	r io.Reader
}

func (r lrs) Read(p []byte) (int, error) {
	return r.r.Read(p)
}

// withSpeedLimit 给原有的ReadSeeker加上限速
func withSpeedLimit(rs io.ReadSeeker, speed int) io.ReadSeeker {
	bucket := ratelimit.NewBucketWithRate(float64(speed), int64(speed))
	lrs := lrs{rs, ratelimit.Reader(rs, bucket)}
	return lrs
}

// AddFile 新增文件记录
func (fs *FileSystem) AddFile(ctx context.Context, parent *model.Folder) (*model.File, error) {
	file := ctx.Value(FileHeaderCtx).(FileHeader)
	filePath := ctx.Value(SavePathCtx).(string)

	newFile := model.File{
		Name:       file.GetFileName(),
		SourceName: filePath,
		UserID:     fs.User.ID,
		Size:       file.GetSize(),
		FolderID:   parent.ID,
		PolicyID:   fs.User.Policy.ID,
		Dir:        parent.PositionAbsolute,
	}
	_, err := newFile.Create()

	if err != nil {
		return nil, err
	}

	return &newFile, nil
}

// GetDownloadContent 获取用于下载的文件流
func (fs *FileSystem) GetDownloadContent(ctx context.Context, path string) (io.ReadSeeker, error) {
	// 获取原始文件流
	rs, err := fs.GetContent(ctx, path)
	if err != nil {
		return nil, err
	}

	// 如果用户组有速度限制，就返回限制流速的ReaderSeeker
	if fs.User.Group.SpeedLimit != 0 {
		return withSpeedLimit(rs, fs.User.Group.SpeedLimit), nil
	}
	// 否则返回原始流
	return rs, nil

}

// GetContent 获取文件内容，path为虚拟路径
// TODO:测试
func (fs *FileSystem) GetContent(ctx context.Context, path string) (io.ReadSeeker, error) {
	// 触发`下载前`钩子
	err := fs.Trigger(ctx, fs.BeforeFileDownload)
	if err != nil {
		util.Log().Debug("BeforeFileDownload 钩子执行失败，%s", err)
		return nil, err
	}

	// 找到文件
	exist, file := fs.IsFileExist(path)
	if !exist {
		return nil, ErrObjectNotExist
	}
	fs.Target = &file

	// 将当前存储策略重设为文件使用的
	fs.Policy = file.GetPolicy()
	err = fs.dispatchHandler()
	if err != nil {
		return nil, err
	}

	// 获取文件流
	rs, err := fs.Handler.Get(ctx, file.SourceName)
	if err != nil {
		return nil, ErrIO.WithError(err)
	}

	return rs, nil
}
