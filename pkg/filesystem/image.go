package filesystem

import (
	"context"
	"fmt"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/thumb"
	"github.com/HFO4/cloudreve/pkg/util"
)

/* ================
     图像处理相关
   ================
*/

// HandledExtension 可以生成缩略图的文件扩展名
var HandledExtension = []string{"jpg", "jpeg", "png", "gif"}

// GenerateThumbnail 尝试为本地策略文件生成缩略图并获取图像原始大小
func (fs *FileSystem) GenerateThumbnail(ctx context.Context, file *model.File) {
	// 判断是否可以生成缩略图
	if !IsInExtensionList(HandledExtension, file.Name) {
		return
	}

	// 新建上下文
	newCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 获取文件数据
	source, err := fs.Handler.Get(newCtx, file.SourceName)
	if err != nil {
		return
	}

	image, err := thumb.NewThumbFromFile(source, file.Name)
	if err != nil {
		util.Log().Warning("生成缩略图时无法解析[%s]图像数据：%s", file.SourceName, err)
		return
	}

	// 获取原始图像尺寸
	w, h := image.GetSize()

	// 生成缩略图
	image.GetThumb(fs.GenerateThumbnailSize(w, h))
	// 保存到文件
	err = image.Save(file.SourceName + "._thumb")
	if err != nil {
		util.Log().Warning("无法保存缩略图：%s", file.SourceName, err)
		return
	}

	// 更新文件的图像信息
	err = file.UpdatePicInfo(fmt.Sprintf("%d,%d", w, h))

	// 失败时删除缩略图文件
	if err != nil {
		_, _ = fs.Handler.Delete(newCtx, []string{file.SourceName + "._thumb"})
	}
}

// GenerateThumbnailSize 获取要生成的缩略图的尺寸
func (fs *FileSystem) GenerateThumbnailSize(w, h int) (uint, uint) {
	return 230, 200
}
