package callback

import (
	"context"
	"github.com/HFO4/cloudreve/pkg/filesystem"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/filesystem/local"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/gin-gonic/gin"
)

// RemoteUploadCallbackService 远程存储上传回调请求服务
type RemoteUploadCallbackService struct {
	Data serializer.RemoteUploadCallback `json:"data" binding:"required"`
}

// Process 处理远程策略上传结果回调
func (service *RemoteUploadCallbackService) Process(c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}
	defer fs.Recycle()

	// 获取回调会话
	callbackSessionRaw, ok := c.Get("callbackSession")
	if !ok {
		return serializer.Err(serializer.CodeInternalSetting, "找不到回调会话", nil)
	}
	callbackSession := callbackSessionRaw.(*serializer.UploadSession)

	// 获取父目录
	exist, parentFolder := fs.IsPathExist(callbackSession.VirtualPath)
	if !exist {
		return serializer.Err(serializer.CodeParamErr, "指定目录不存在", nil)
	}

	// 创建文件头
	fileHeader := local.FileStream{
		Size:        service.Data.Size,
		VirtualPath: callbackSession.VirtualPath,
		Name:        service.Data.Name,
	}

	// 生成上下文
	ctx := context.WithValue(context.Background(), fsctx.FileHeaderCtx, fileHeader)
	ctx = context.WithValue(ctx, fsctx.SavePathCtx, service.Data.SourceName)

	// 添加钩子
	fs.Use("BeforeAddFile", filesystem.HookValidateFile)
	fs.Use("BeforeAddFile", filesystem.HookValidateCapacity)
	fs.Use("AfterValidateFailed", filesystem.HookGiveBackCapacity)

	// 向数据库中添加文件
	file, err := fs.AddFile(ctx, parentFolder)
	if err != nil {
		return serializer.Err(serializer.CodeUploadFailed, err.Error(), err)
	}

	// 如果是图片，则更新图片信息
	if service.Data.PicInfo != "" {
		if err := file.UpdatePicInfo(service.Data.PicInfo); err != nil {
			util.Log().Debug("无法更新回调文件的图片信息：%s", err)
		}
	}

	return serializer.Response{
		Code: 0,
	}
}
