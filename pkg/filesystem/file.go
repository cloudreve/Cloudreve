package filesystem

import (
	"context"
	"errors"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/serializer"
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

// Download 处理下载文件请求，path为虚拟路径
// TODO:测试
func (fs *FileSystem) Download(ctx context.Context, path string) (io.ReadCloser, error) {
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

	// 将当前存储策略重设为文件使用的
	fs.Policy = file.GetPolicy()
	err = fs.dispatchHandler()
	if err != nil {
		return nil, err
	}

	return nil, serializer.NewError(serializer.CodeEncryptError, "人都的", errors.New("不是人都的"))
}
