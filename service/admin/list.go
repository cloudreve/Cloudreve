package admin

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/serializer"
)

// GroupList 获取用户组列表
func (service *NoParamService) GroupList() serializer.Response {
	var res []model.Group
	model.DB.Model(&model.Group{}).Find(&res)
	return serializer.Response{Data: res}
}
