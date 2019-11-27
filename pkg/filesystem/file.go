package filesystem

import (
	"context"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/util"
	"io"
)

/* ============
	 文件相关
   ============
*/

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
