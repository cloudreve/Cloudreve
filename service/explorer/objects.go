package explorer

import (
	"context"
	"github.com/HFO4/cloudreve/pkg/filesystem"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/gin-gonic/gin"
)

// ItemService 处理多文件/目录相关服务
type ItemService struct {
	Items []string `json:"items" binding:"exists"`
	Dirs  []string `json:"dirs" binding:"exists"`
}

// Delete 删除对象
func (service *ItemService) Delete(ctx context.Context, c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}

	// 删除对象
	err = fs.Delete(ctx, service.Dirs, service.Items)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	return serializer.Response{
		Code: 0,
	}

}
