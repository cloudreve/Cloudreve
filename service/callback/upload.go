package callback

import (
	"context"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/filesystem/local"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/gin-gonic/gin"
)

// CallbackProcessService 上传请求回调正文接口
type CallbackProcessService interface {
	GetBody(*serializer.UploadSession) serializer.UploadCallback
}

// RemoteUploadCallbackService 远程存储上传回调请求服务
type RemoteUploadCallbackService struct {
	Data serializer.UploadCallback `json:"data" binding:"required"`
}

// GetBody 返回回调正文
func (service RemoteUploadCallbackService) GetBody(session *serializer.UploadSession) serializer.UploadCallback {
	return service.Data
}

// UploadCallbackService OOS/七牛云存储上传回调请求服务
type UploadCallbackService struct {
	Name       string `json:"name"`
	SourceName string `json:"source_name"`
	PicInfo    string `json:"pic_info"`
	Size       uint64 `json:"size"`
}

// UpyunCallbackService 又拍云上传回调请求服务
type UpyunCallbackService struct {
	Code       int    `form:"code" binding:"required"`
	Message    string `form:"message" binding:"required"`
	SourceName string `form:"url" binding:"required"`
	Width      string `form:"image-width"`
	Height     string `form:"image-height"`
	Size       uint64 `form:"file_size"`
}

// GetBody 返回回调正文
func (service UpyunCallbackService) GetBody(session *serializer.UploadSession) serializer.UploadCallback {
	res := serializer.UploadCallback{
		Name:       session.Name,
		SourceName: service.SourceName,
		Size:       service.Size,
	}
	if service.Width != "" {
		res.PicInfo = service.Width + "," + service.Height
	}

	return res
}

// GetBody 返回回调正文
func (service UploadCallbackService) GetBody(session *serializer.UploadSession) serializer.UploadCallback {
	return serializer.UploadCallback{
		Name:       service.Name,
		SourceName: service.SourceName,
		PicInfo:    service.PicInfo,
		Size:       service.Size,
	}
}

// ProcessCallback 处理上传结果回调
func ProcessCallback(service CallbackProcessService, c *gin.Context) serializer.Response {
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

	// 获取回调正文
	callbackBody := service.GetBody(callbackSession)

	// 重新指向上传策略
	policy, err := model.GetPolicyByID(callbackSession.PolicyID)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}
	fs.Policy = &policy
	fs.User.Policy = policy
	err = fs.DispatchHandler()
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}

	// 获取父目录
	exist, parentFolder := fs.IsPathExist(callbackSession.VirtualPath)
	if !exist {
		return serializer.Err(serializer.CodeParamErr, "指定目录不存在", nil)
	}

	// 创建文件头
	fileHeader := local.FileStream{
		Size:        callbackBody.Size,
		VirtualPath: callbackSession.VirtualPath,
		Name:        callbackBody.Name,
	}

	// 生成上下文
	ctx := context.WithValue(context.Background(), fsctx.FileHeaderCtx, fileHeader)
	ctx = context.WithValue(ctx, fsctx.SavePathCtx, callbackBody.SourceName)

	// 添加钩子
	fs.Use("BeforeAddFile", filesystem.HookValidateFile)
	fs.Use("BeforeAddFile", filesystem.HookValidateCapacity)
	fs.Use("AfterValidateFailed", filesystem.HookGiveBackCapacity)
	fs.Use("AfterValidateFailed", filesystem.HookDeleteTempFile)
	fs.Use("BeforeAddFileFailed", filesystem.HookDeleteTempFile)

	// 向数据库中添加文件
	file, err := fs.AddFile(ctx, parentFolder)
	if err != nil {
		return serializer.Err(serializer.CodeUploadFailed, err.Error(), err)
	}

	// 如果是图片，则更新图片信息
	if callbackBody.PicInfo != "" {
		if err := file.UpdatePicInfo(callbackBody.PicInfo); err != nil {
			util.Log().Debug("无法更新回调文件的图片信息：%s", err)
		}
	}

	return serializer.Response{
		Code: 0,
	}
}
