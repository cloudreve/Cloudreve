package filesystem

import (
	"context"
	"errors"
	"io"
	"sync"

	"runtime"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/response"
	"github.com/cloudreve/Cloudreve/v3/pkg/thumb"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

/* ================
     图像处理相关
   ================
*/

// GetThumb 获取文件的缩略图
func (fs *FileSystem) GetThumb(ctx context.Context, id uint) (*response.ContentResponse, error) {
	// 根据 ID 查找文件
	err := fs.resetFileIDIfNotExist(ctx, id)
	if err != nil {
		return &response.ContentResponse{
			Redirect: false,
		}, ErrObjectNotExist
	}

	w, h := fs.GenerateThumbnailSize(0, 0)
	ctx = context.WithValue(ctx, fsctx.ThumbSizeCtx, [2]uint{w, h})
	ctx = context.WithValue(ctx, fsctx.FileModelCtx, fs.FileTarget[0])
	res, err := fs.Handler.Thumb(ctx, &fs.FileTarget[0])
	if errors.Is(err, driver.ErrorThumbNotExist) {
		// Regenerate thumb if the thumb is not initialized yet
		fs.GenerateThumbnail(ctx, &fs.FileTarget[0])
		res, err = fs.Handler.Thumb(ctx, &fs.FileTarget[0])
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
		maxWorker := model.GetIntSetting("thumb_max_task_count", -1)
		if maxWorker <= 0 {
			maxWorker = runtime.GOMAXPROCS(0)
		}
		thumbPool = &Pool{
			worker: make(chan int, maxWorker),
		}
		util.Log().Debug("Initialize thumbnails task queue with: WorkerNum = %d", maxWorker)
	})
	return thumbPool
}
func (pool *Pool) addWorker() {
	pool.worker <- 1
	util.Log().Debug("Worker added to thumbnails task queue.")
}
func (pool *Pool) releaseWorker() {
	util.Log().Debug("Worker released from thumbnails task queue.")
	<-pool.worker
}

// GenerateThumbnail generates thumb for given file, upload the thumb file back with given suffix
func (fs *FileSystem) GenerateThumbnail(ctx context.Context, file *model.File) {
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

	r, w := io.Pipe()
	defer w.Close()

	errChan := make(chan error, 1)
	go func(errChan chan error) {
		errChan <- fs.Handler.Put(newCtx, &fsctx.FileStream{
			Mode:     fsctx.Overwrite,
			File:     io.NopCloser(r),
			Seeker:   nil,
			SavePath: file.SourceName + model.GetSettingByNameWithDefault("thumb_file_suffix", "._thumb"),
		})
	}(errChan)

	if err = thumb.Generators.Generate(source, w, file.Name, model.GetSettingByNames(
		"thumb_width",
		"thumb_height",
		"thumb_builtin_enabled",
		"thumb_vips_enabled",
		"thumb_ffmpeg_enabled",
		"thumb_vips_path",
		"thumb_ffmpeg_path",
	)); err != nil {
		util.Log().Warning("Failed to generate thumb for %s: %s", file.Name, err)
		_ = updateThumbStatus(file, model.ThumbStatusNotAvailable)
		w.Close()
		<-errChan
		return
	}

	w.Close()
	if err = <-errChan; err != nil {
		util.Log().Warning("Failed to save thumb for %s: %s", file.Name, err)
		return
	}

	if model.IsTrueVal(model.GetSettingByName("thumb_gc_after_gen")) {
		util.Log().Debug("GenerateThumbnail runtime.GC")
		runtime.GC()
	}

	// Mark this file as thumb available
	err = updateThumbStatus(file, model.ThumbStatusExist)

	// 失败时删除缩略图文件
	if err != nil {
		_, _ = fs.Handler.Delete(newCtx, []string{file.SourceName + model.GetSettingByNameWithDefault("thumb_file_suffix", "._thumb")})
	}
}

// GenerateThumbnailSize 获取要生成的缩略图的尺寸
func (fs *FileSystem) GenerateThumbnailSize(w, h int) (uint, uint) {
	return uint(model.GetIntSetting("thumb_width", 400)), uint(model.GetIntSetting("thumb_height", 300))
}

func updateThumbStatus(file *model.File, status string) error {
	if file.Model.ID > 0 {
		return file.UpdateMetadata(map[string]string{
			model.ThumbStatusMetadataKey: status,
		})
	} else {
		file.MetadataSerialized[model.ThumbStatusMetadataKey] = status
	}

	return nil
}
