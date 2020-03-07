package admin

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/serializer"
)

// Users 列出用户
func (service *AdminListService) Users() serializer.Response {
	var res []model.User
	total := 0

	tx := model.DB.Model(&model.User{})
	if service.OrderBy != "" {
		tx = tx.Order(service.OrderBy)
	}

	for k, v := range service.Conditions {
		tx = tx.Where(k+" = ?", v)
	}

	// 计算总数用于分页
	tx.Count(&total)

	// 查询记录
	tx.Set("gorm:auto_preload", true).Limit(service.PageSize).Offset((service.Page - 1) * service.PageSize).Find(&res)

	// 补齐缺失用户组

	return serializer.Response{Data: map[string]interface{}{
		"total": total,
		"items": res,
	}}
}
