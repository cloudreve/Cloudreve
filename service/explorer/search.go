package explorer

import (
	"context"
	"github.com/HFO4/cloudreve/pkg/filesystem"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/gin-gonic/gin"
)

// ItemSearchService 文件搜索服务
type ItemSearchService struct {
	Type     string `uri:"type" binding:"required"`
	Keywords string `uri:"keywords" binding:"required"`
}

// Search 执行搜索
func (service *ItemSearchService) Search(c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}
	defer fs.Recycle()

	switch service.Type {
	case "keywords":
		return service.SearchKeywords(c, "%"+service.Keywords+"%", fs)
	default:
		return serializer.ParamErr("未知搜索类型", nil)
	}
}

// SearchKeywords 根据关键字搜索文件
func (service *ItemSearchService) SearchKeywords(c *gin.Context, keywords string, fs *filesystem.FileSystem) serializer.Response {
	// 上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 获取子项目
	objects, err := fs.Search(ctx, keywords)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	return serializer.Response{
		Code: 0,
		Data: map[string]interface{}{
			"parent":  0,
			"objects": objects,
		},
	}
}
