package filesystem

import (
	"context"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/gin-gonic/gin"
	"path/filepath"
)

/* ================
	 上传处理相关
   ================
*/

// Upload 上传文件
func (fs *FileSystem) Upload(ctx context.Context, file FileHeader) (err error) {
	ctx = context.WithValue(ctx, fsctx.FileHeaderCtx, file)

	// 上传前的钩子
	err = fs.Trigger(ctx, fs.BeforeUpload)
	if err != nil {
		return err
	}

	// 生成文件名和路径,
	var savePath string
	// 如果是更新操作就从上下文中获取
	if originFile, ok := ctx.Value(fsctx.FileModelCtx).(model.File); ok {
		savePath = originFile.SourceName
	} else {
		savePath = fs.GenerateSavePath(ctx, file)
	}
	ctx = context.WithValue(ctx, fsctx.SavePathCtx, savePath)

	// 处理客户端未完成上传时，关闭连接
	go fs.CancelUpload(ctx, savePath, file)

	// 保存文件
	err = fs.Handler.Put(ctx, file, savePath, file.GetSize())
	if err != nil {
		return err
	}

	// 上传完成后的钩子
	err = fs.Trigger(ctx, fs.AfterUpload)

	if err != nil {
		// 上传完成后续处理失败
		followUpErr := fs.Trigger(ctx, fs.AfterValidateFailed)
		// 失败后再失败...
		if followUpErr != nil {
			util.Log().Debug("AfterValidateFailed 钩子执行失败，%s", followUpErr)
		}

		return err
	}

	util.Log().Info(
		"新文件PUT:%s , 大小:%d, 上传者:%s",
		file.GetFileName(),
		file.GetSize(),
		fs.User.Nick,
	)

	return nil
}

// GenerateSavePath 生成要存放文件的路径
// TODO 完善测试
func (fs *FileSystem) GenerateSavePath(ctx context.Context, file FileHeader) string {
	if fs.User.Model.ID != 0 {
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

	// 匿名文件系统使用空上传策略生成路径
	nilPolicy := model.Policy{}
	return filepath.Join(
		nilPolicy.GeneratePath(
			0,
			"",
		),
		nilPolicy.GenerateFileName(
			0,
			"",
		),
	)
}

// CancelUpload 监测客户端取消上传
func (fs *FileSystem) CancelUpload(ctx context.Context, path string, file FileHeader) {
	var reqContext context.Context
	if ginCtx, ok := ctx.Value(fsctx.GinCtx).(*gin.Context); ok {
		reqContext = ginCtx.Request.Context()
	} else {
		reqContext = ctx.Value(fsctx.HTTPCtx).(context.Context)
	}

	select {
	case <-reqContext.Done():
		select {
		case <-ctx.Done():
			// 客户端正常关闭，不执行操作
		default:
			// 客户端取消上传，删除临时文件
			util.Log().Debug("客户端取消上传")
			if fs.AfterUploadCanceled == nil {
				return
			}
			ctx = context.WithValue(ctx, fsctx.SavePathCtx, path)
			err := fs.Trigger(ctx, fs.AfterUploadCanceled)
			if err != nil {
				util.Log().Debug("执行 AfterUploadCanceled 钩子出错，%s", err)
			}
		}

	}
}
