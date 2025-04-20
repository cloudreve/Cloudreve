package admin

import (
	"github.com/cloudreve/Cloudreve/v4/pkg/serializer"
)

// AdminListService 仪表盘列条目服务
type (
	AdminListService struct {
		Page           int               `json:"page" binding:"min=1"`
		PageSize       int               `json:"page_size" binding:"min=1,required"`
		OrderBy        string            `json:"order_by"`
		OrderDirection string            `json:"order_direction"`
		Conditions     map[string]string `json:"conditions"`
		Searches       map[string]string `json:"searches"`
	}
	AdminListServiceParamsCtx struct{}
)

// GroupList 获取用户组列表
func (service *NoParamService) GroupList() serializer.Response {
	//var res []model.Group
	//model.DB.Model(&model.Group{}).Find(&res)
	//return serializer.Response{Data: res}
	return serializer.Response{}
}
