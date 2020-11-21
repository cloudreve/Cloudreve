package explorer

import (
	"fmt"
	"strings"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/gin-gonic/gin"
)

// FilterTagCreateService 文件分类标签创建服务
type FilterTagCreateService struct {
	Expression string `json:"expression" binding:"required,min=1,max=65535"`
	Icon       string `json:"icon" binding:"required,min=1,max=255"`
	Name       string `json:"name" binding:"required,min=1,max=255"`
	Color      string `json:"color" binding:"hexcolor|rgb|rgba|hsl"`
}

// LinkTagCreateService 目录快捷方式标签创建服务
type LinkTagCreateService struct {
	Path string `json:"path" binding:"required,min=1,max=65535"`
	Name string `json:"name" binding:"required,min=1,max=255"`
}

// TagService 标签服务
type TagService struct {
}

// Delete 删除标签
func (service *TagService) Delete(c *gin.Context, user *model.User) serializer.Response {
	id, _ := c.Get("object_id")
	if err := model.DeleteTagByID(id.(uint), user.ID); err != nil {
		return serializer.Err(serializer.CodeDBError, "删除失败", err)
	}
	return serializer.Response{}
}

// Create 创建标签
func (service *LinkTagCreateService) Create(c *gin.Context, user *model.User) serializer.Response {
	// 创建标签
	tag := model.Tag{
		Name:       service.Name,
		Icon:       "FolderHeartOutline",
		Type:       model.DirectoryLinkType,
		Expression: service.Path,
		UserID:     user.ID,
	}
	id, err := tag.Create()
	if err != nil {
		return serializer.Err(serializer.CodeDBError, "标签创建失败", err)
	}

	return serializer.Response{
		Data: hashid.HashID(id, hashid.TagID),
	}
}

// Create 创建标签
func (service *FilterTagCreateService) Create(c *gin.Context, user *model.User) serializer.Response {
	// 分割表达式，将通配符转换为SQL内的%
	expressions := strings.Split(service.Expression, "\n")
	for i := 0; i < len(expressions); i++ {
		expressions[i] = strings.ReplaceAll(expressions[i], "*", "%")
		if expressions[i] == "" {
			return serializer.ParamErr(fmt.Sprintf("第 %d 行包含空的匹配表达式", i+1), nil)
		}
	}

	// 创建标签
	tag := model.Tag{
		Name:       service.Name,
		Icon:       service.Icon,
		Color:      service.Color,
		Type:       model.FileTagType,
		Expression: strings.Join(expressions, "\n"),
		UserID:     user.ID,
	}
	id, err := tag.Create()
	if err != nil {
		return serializer.Err(serializer.CodeDBError, "标签创建失败", err)
	}

	return serializer.Response{
		Data: hashid.HashID(id, hashid.TagID),
	}
}
