package filesystem

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"runtime"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/response"
	"github.com/cloudreve/Cloudreve/v3/pkg/thumb"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

/* ================
     图像处理相关
   ================
*/

// HandledExtension 可以生成缩略图的文件扩展名
var HandledExtension = []string{"jpg", "jpeg", "png", "gif"}

// GetThumb 获取文件的缩略图
func (fs *FileSystem) GetThumb(ctx context.Context, id uint) (*response.ContentResponse, error) {
	// 根据 ID 查找文件
	err := fs.resetFileIDIfNotExist(ctx, id)
	if err != nil || fs.FileTarget[0].PicInfo == "" {
		return &response.ContentResponse{
			Redirect: false,
		}, ErrObjectNotExist
	}

	w, h := fs.GenerateThumbnailSize(0, 0)
	ctx = context.WithValue(ctx, fsctx.ThumbSizeCtx, [2]uint{w, h})
	ctx = context.WithValue(ctx, fsctx.FileModelCtx, fs.FileTarget[0])
	res, err := fs.Handler.Thumb(ctx, fs.FileTarget[0].SourceName)

	// 本地存储策略出错时重新生成缩略图
	if err != nil && fs.Policy.Type == "local" {
		fs.GenerateThumbnail(ctx, &fs.FileTarget[0])
		res, err = fs.Handler.Thumb(ctx, fs.FileTarget[0].SourceName)
	}

	if err == nil && conf.SystemConfig.Mode == "master" {
		res.MaxAge = model.GetIntSetting("preview_timeout", 60)
	}

	return res, err
}

// thumbPool 要使用的任务池
var thumbPool *Pool
var once sync.Once

// Pool 带有最大配额的任务池
type Pool struct {
	// 容量
	worker chan int
}

// Init 初始化任务池
func getThumbWorker() *Pool {
	once.Do(func() {
		maxWorker := conf.ThumbConfig.MaxTaskCount
		if maxWorker <= 0 {
			maxWorker = runtime.GOMAXPROCS(0)
		}
		thumbPool = &Pool{
			worker: make(chan int, maxWorker),
		}
		util.Log().Debug("初始化Thumb任务队列，WorkerNum = %d", maxWorker)
	})
	return thumbPool
}
func (pool *Pool) addWorker() {
	pool.worker <- 1
	util.Log().Debug("Thumb任务队列，addWorker")
}
func (pool *Pool) releaseWorker() {
	util.Log().Debug("Thumb任务队列，releaseWorker")
	<-pool.worker
}

// GenerateThumbnail 尝试为本地策略文件生成缩略图并获取图像原始大小
// TODO 失败时，如果之前还有图像信息，则清除
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
	defer source.Close()
	getThumbWorker().addWorker()
	defer getThumbWorker().releaseWorker()

	image, err := thumb.NewThumbFromFile(source, file.Name)
	if err != nil {
		util.Log().Warning("生成缩略图时无法解析 [%s] 图像数据：%s", file.SourceName, err)
		return
	}

	// 获取原始图像尺寸
	w, h := image.GetSize()

	// 生成缩略图
	image.GetThumb(fs.GenerateThumbnailSize(w, h))
	// 保存到文件
	err = image.Save(util.RelativePath(file.SourceName + conf.ThumbConfig.FileSuffix))
	image = nil
	if conf.ThumbConfig.GCAfterGen {
		util.Log().Debug("GenerateThumbnail runtime.GC")
		runtime.GC()
	}

	if err != nil {
		util.Log().Warning("无法保存缩略图：%s", err)
		return
	}

	// 更新文件的图像信息
	if file.Model.ID > 0 {
		err = file.UpdatePicInfo(fmt.Sprintf("%d,%d", w, h))
	} else {
		file.PicInfo = fmt.Sprintf("%d,%d", w, h)
	}

	// 失败时删除缩略图文件
	if err != nil {
		_, _ = fs.Handler.Delete(newCtx, []string{file.SourceName + conf.ThumbConfig.FileSuffix})
	}
}

// GenerateThumbnailSize 获取要生成的缩略图的尺寸
func (fs *FileSystem) GenerateThumbnailSize(w, h int) (uint, uint) {
	if conf.SystemConfig.Mode == "master" {
		options := model.GetSettingByNames("thumb_width", "thumb_height")
		w, _ := strconv.ParseUint(options["thumb_width"], 10, 32)
		h, _ := strconv.ParseUint(options["thumb_height"], 10, 32)
		return uint(w), uint(h)
	}
	return conf.ThumbConfig.MaxWidth, conf.ThumbConfig.MaxHeight
}
