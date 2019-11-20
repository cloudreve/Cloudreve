package filesystem

import (
	"context"
	model "github.com/HFO4/cloudreve/models"
)

/* ===============
	图像处理相关
   ===============
*/

// HandledExtension 可以生成缩略图的文件扩展名
var HandledExtension = []string{}

// GenerateThumbnail 尝试为文件生成缩略图
func (fs *FileSystem) GenerateThumbnail(ctx context.Context, file *model.File) {
	// TODO
}
