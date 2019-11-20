package filesystem

import (
	"context"
	model "github.com/HFO4/cloudreve/models"
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
