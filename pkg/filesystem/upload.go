package filesystem

import (
	"context"
	"fmt"
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

	util.Log().Info("新文件上传:%s , 大小:%d, 上传者:%s", file.GetFileName(), file.GetSize(), fs.User.Nick)

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
		fmt.Println("正常")
	case <-ginCtx.Request.Context().Done():
		// 客户端取消了上传
		fmt.Println("取消")
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
