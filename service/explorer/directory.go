package explorer

import (
	"context"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/gin-gonic/gin"
)

// DirectoryCreateService 创建新目录服务
type DirectoryCreateService struct {
	Path string `form:"newPath" json:"newPath" binding:"required,min=1,max=65535"`
}

// CreateDirectory 创建目录
func (service *DirectoryCreateService) CreateDirectory(c *gin.Context) serializer.Response {
	// 创建文件系统
	user, _ := c.Get("user")
	fs, err := filesystem.NewFileSystem(user.(*model.User))
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}

	// 上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建目录
	err = fs.CreateDirectory(ctx, service.Path)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFolderFailed, err.Error(), err)
	}

	return serializer.Response{
		Code: 0,
	}

}
