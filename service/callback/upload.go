package callback

import (
	"context"
	"fmt"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"strings"

	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/cos"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/onedrive"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/s3"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/gin-gonic/gin"
)

// CallbackProcessService 上传请求回调正文接口
type CallbackProcessService interface {
	GetBody() serializer.UploadCallback
}

// RemoteUploadCallbackService 远程存储上传回调请求服务
type RemoteUploadCallbackService struct {
	Data serializer.UploadCallback `json:"data" binding:"required"`
}

// GetBody 返回回调正文
func (service RemoteUploadCallbackService) GetBody() serializer.UploadCallback {
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

// OneDriveCallback OneDrive 客户端回调正文
type OneDriveCallback struct {
	Meta *onedrive.FileInfo
}

// COSCallback COS 客户端回调正文
type COSCallback struct {
	Bucket string `form:"bucket"`
	Etag   string `form:"etag"`
}

// S3Callback S3 客户端回调正文
type S3Callback struct {
}

// GetBody 返回回调正文
func (service UpyunCallbackService) GetBody() serializer.UploadCallback {
	res := serializer.UploadCallback{}
	if service.Width != "" {
		res.PicInfo = service.Width + "," + service.Height
	}

	return res
}

// GetBody 返回回调正文
func (service UploadCallbackService) GetBody() serializer.UploadCallback {
	return serializer.UploadCallback{
		PicInfo: service.PicInfo,
	}
}

// GetBody 返回回调正文
func (service OneDriveCallback) GetBody() serializer.UploadCallback {
	var picInfo = "0,0"
	if service.Meta.Image.Width != 0 {
		picInfo = fmt.Sprintf("%d,%d", service.Meta.Image.Width, service.Meta.Image.Height)
	}
	return serializer.UploadCallback{
		PicInfo: picInfo,
	}
}

// GetBody 返回回调正文
func (service COSCallback) GetBody() serializer.UploadCallback {
	return serializer.UploadCallback{
		PicInfo: "",
	}
}

// GetBody 返回回调正文
func (service S3Callback) GetBody() serializer.UploadCallback {
	return serializer.UploadCallback{
		PicInfo: "",
	}
}

// ProcessCallback 处理上传结果回调
func ProcessCallback(service CallbackProcessService, c *gin.Context) serializer.Response {
	callbackBody := service.GetBody()

	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromCallback(c)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, err.Error(), err)
	}
	defer fs.Recycle()

	// 获取上传会话
	uploadSession := c.MustGet(filesystem.UploadSessionCtx).(*serializer.UploadSession)

	// 查找上传会话创建的占位文件
	file, err := model.GetFilesByUploadSession(uploadSession.Key, fs.User.ID)
	if err != nil {
		return serializer.Err(serializer.CodeUploadSessionExpired, "LocalUpload session file placeholder not exist", err)
	}

	fileData := fsctx.FileStream{
		Size:         uploadSession.Size,
		Name:         uploadSession.Name,
		VirtualPath:  uploadSession.VirtualPath,
		SavePath:     uploadSession.SavePath,
		Mode:         fsctx.Nop,
		Model:        file,
		LastModified: uploadSession.LastModified,
	}

	// 占位符未扣除容量需要校验和扣除
	if !fs.Policy.IsUploadPlaceholderWithSize() {
		fs.Use("AfterUpload", filesystem.HookValidateCapacity)
		fs.Use("AfterUpload", filesystem.HookChunkUploaded)
	}

	fs.Use("AfterUpload", filesystem.HookPopPlaceholderToFile(callbackBody.PicInfo))
	fs.Use("AfterValidateFailed", filesystem.HookDeleteTempFile)
	err = fs.Upload(context.Background(), &fileData)
	if err != nil {
		return serializer.Err(serializer.CodeUploadFailed, err.Error(), err)
	}

	return serializer.Response{}
}

// PreProcess 对OneDrive客户端回调进行预处理验证
func (service *OneDriveCallback) PreProcess(c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromCallback(c)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	// 获取回调会话
	uploadSession := c.MustGet(filesystem.UploadSessionCtx).(*serializer.UploadSession)

	// 获取文件信息
	info, err := fs.Handler.(onedrive.Driver).Client.Meta(context.Background(), "", uploadSession.SavePath)
	if err != nil {
		return serializer.Err(serializer.CodeQueryMetaFailed, "", err)
	}

	// 验证与回调会话中是否一致
	actualPath := strings.TrimPrefix(uploadSession.SavePath, "/")
	isSizeCheckFailed := uploadSession.Size != info.Size

	// SharePoint 会对 Office 文档增加 meta data 导致文件大小不一致，这里增加 1 MB 宽容
	// See: https://github.com/OneDrive/onedrive-api-docs/issues/935
	if (strings.Contains(fs.Policy.OptionsSerialized.OdDriver, "sharepoint.com") || strings.Contains(fs.Policy.OptionsSerialized.OdDriver, "sharepoint.cn")) && isSizeCheckFailed && (info.Size > uploadSession.Size) && (info.Size-uploadSession.Size <= 1048576) {
		isSizeCheckFailed = false
	}

	if isSizeCheckFailed || !strings.EqualFold(info.GetSourcePath(), actualPath) {
		fs.Handler.(onedrive.Driver).Client.Delete(context.Background(), []string{info.GetSourcePath()})
		return serializer.Err(serializer.CodeMetaMismatch, "", err)
	}
	service.Meta = info
	return ProcessCallback(service, c)
}

// PreProcess 对COS客户端回调进行预处理
func (service *COSCallback) PreProcess(c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromCallback(c)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	// 获取回调会话
	uploadSession := c.MustGet(filesystem.UploadSessionCtx).(*serializer.UploadSession)

	// 获取文件信息
	info, err := fs.Handler.(cos.Driver).Meta(context.Background(), uploadSession.SavePath)
	if err != nil {
		return serializer.Err(serializer.CodeMetaMismatch, "", err)
	}

	// 验证实际文件信息与回调会话中是否一致
	if uploadSession.Size != info.Size || uploadSession.Key != info.CallbackKey {
		return serializer.Err(serializer.CodeMetaMismatch, "", err)
	}

	return ProcessCallback(service, c)
}

// PreProcess 对S3客户端回调进行预处理
func (service *S3Callback) PreProcess(c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromCallback(c)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	// 获取回调会话
	uploadSession := c.MustGet(filesystem.UploadSessionCtx).(*serializer.UploadSession)

	// 获取文件信息
	info, err := fs.Handler.(*s3.Driver).Meta(context.Background(), uploadSession.SavePath)
	if err != nil {
		return serializer.Err(serializer.CodeMetaMismatch, "", err)
	}

	// 验证实际文件信息与回调会话中是否一致
	if uploadSession.Size != info.Size {
		return serializer.Err(serializer.CodeMetaMismatch, "", err)
	}

	return ProcessCallback(service, c)
}

// PreProcess 对OneDrive客户端回调进行预处理验证
func (service *UploadCallbackService) PreProcess(c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromCallback(c)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	// 获取回调会话
	uploadSession := c.MustGet(filesystem.UploadSessionCtx).(*serializer.UploadSession)

	// 验证文件大小
	if uploadSession.Size != service.Size {
		fs.Handler.Delete(context.Background(), []string{uploadSession.SavePath})
		return serializer.Err(serializer.CodeMetaMismatch, "", err)
	}

	return ProcessCallback(service, c)
}
