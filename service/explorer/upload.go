package explorer

import (
	"context"

	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/gin-gonic/gin"
)

// UploadSessionService 获取上传凭证服务
type UploadSessionService struct {
	Path     string `json:"path" binding:"required"`
	Size     uint64 `json:"size" binding:"min=0"`
	Name     string `json:"name" binding:"required"`
	PolicyID string `json:"policy_id" binding:"required"`
}

// Create 创建新的上传会话
func (service *UploadSessionService) Create(ctx context.Context, c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}

	// 取得存储策略的ID
	rawID, err := hashid.DecodeHashID(service.PolicyID, hashid.PolicyID)
	if err != nil {
		return serializer.Err(serializer.CodeNotFound, "存储策略不存在", err)
	}

	if fs.Policy.ID != rawID {
		return serializer.Err(serializer.CodePolicyNotAllowed, "存储策略发生变化，请刷新文件列表并重新添加此任务", nil)
	}

	credential, err := fs.CreateUploadSession(ctx, service.Path, service.Size, service.Name)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	return serializer.Response{
		Code: 0,
		Data: credential,
	}
}
