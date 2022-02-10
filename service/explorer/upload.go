package explorer

import (
	"context"

	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/gin-gonic/gin"
)

// UploadCredentialService 获取上传凭证服务
type UploadCredentialService struct {
	Path string `form:"path" binding:"required"`
	Size uint64 `form:"size" binding:"min=0"`
	Name string `form:"name"`
	Type string `form:"type"`
}

// Get 获取新的上传凭证
func (service *UploadCredentialService) Get(ctx context.Context, c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}

	ctx = context.WithValue(ctx, fsctx.GinCtx, c)
	credential, err := fs.GetUploadToken(ctx, service.Path, service.Size, service.Name)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	return serializer.Response{
		Code: 0,
		Data: credential,
	}
}
