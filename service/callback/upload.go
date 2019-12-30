package callback

import (
	"github.com/HFO4/cloudreve/pkg/filesystem"
	"github.com/HFO4/cloudreve/pkg/serializer"
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

	return serializer.Response{
		Code: 0,
	}
}
