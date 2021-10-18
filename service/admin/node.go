package admin

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cluster"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"strings"
)

// Nodes 列出从机节点
func (service *AdminListService) Nodes() serializer.Response {
	var res []model.Node
	total := 0

	tx := model.DB.Model(&model.Node{})
	if service.OrderBy != "" {
		tx = tx.Order(service.OrderBy)
	}

	for k, v := range service.Conditions {
		tx = tx.Where(k+" = ?", v)
	}

	if len(service.Searches) > 0 {
		search := ""
		for k, v := range service.Searches {
			search += k + " like '%" + v + "%' OR "
		}
		search = strings.TrimSuffix(search, " OR ")
		tx = tx.Where(search)
	}

	// 计算总数用于分页
	tx.Count(&total)

	// 查询记录
	tx.Limit(service.PageSize).Offset((service.Page - 1) * service.PageSize).Find(&res)

	isActive := make(map[uint]bool)
	for i := 0; i < len(res); i++ {
		if node := cluster.Default.GetNodeByID(res[i].ID); node != nil {
			isActive[res[i].ID] = node.IsActive()
		}
	}

	return serializer.Response{Data: map[string]interface{}{
		"total":  total,
		"items":  res,
		"active": isActive,
	}}
}
