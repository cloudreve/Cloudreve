package admin

import (
	"fmt"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
)

// AddGroupService 用户组添加服务
type AddGroupService struct {
	Group model.Group `json:"group" binding:"required"`
}

// GroupService 用户组ID服务
type GroupService struct {
	ID uint `uri:"id" json:"id" binding:"required"`
}

// Get 获取用户组详情
func (service *GroupService) Get() serializer.Response {
	group, err := model.GetGroupByID(service.ID)
	if err != nil {
		return serializer.Err(serializer.CodeNotFound, "存储策略不存在", err)
	}

	return serializer.Response{Data: group}
}

// Delete 删除用户组
func (service *GroupService) Delete() serializer.Response {
	// 查找用户组
	group, err := model.GetGroupByID(service.ID)
	if err != nil {
		return serializer.Err(serializer.CodeNotFound, "用户组不存在", err)
	}

	// 是否为系统用户组
	if group.ID <= 3 {
		return serializer.Err(serializer.CodeNoPermissionErr, "系统用户组无法删除", err)
	}

	// 检查是否有用户使用
	total := 0
	row := model.DB.Model(&model.User{}).Where("group_id = ?", service.ID).
		Select("count(id)").Row()
	row.Scan(&total)
	if total > 0 {
		return serializer.ParamErr(fmt.Sprintf("有 %d 位用户仍属于此用户组，请先删除这些用户或者更改用户组", total), nil)
	}

	model.DB.Delete(&group)

	return serializer.Response{}
}

// Add 添加用户组
func (service *AddGroupService) Add() serializer.Response {
	if service.Group.ID > 0 {
		if err := model.DB.Save(&service.Group).Error; err != nil {
			return serializer.ParamErr("用户组保存失败", err)
		}
	} else {
		if err := model.DB.Create(&service.Group).Error; err != nil {
			return serializer.ParamErr("用户组添加失败", err)
		}
	}

	return serializer.Response{Data: service.Group.ID}
}

// Groups 列出用户组
func (service *AdminListService) Groups() serializer.Response {
	var res []model.Group
	total := 0

	tx := model.DB.Model(&model.Group{})
	if service.OrderBy != "" {
		tx = tx.Order(service.OrderBy)
	}

	for k, v := range service.Conditions {
		tx = tx.Where(k+" = ?", v)
	}

	// 计算总数用于分页
	tx.Count(&total)

	// 查询记录
	tx.Limit(service.PageSize).Offset((service.Page - 1) * service.PageSize).Find(&res)

	// 统计每个用户组的用户总数
	statics := make(map[uint]int, len(res))
	for i := 0; i < len(res); i++ {
		total := 0
		row := model.DB.Model(&model.User{}).Where("group_id = ?", res[i].ID).
			Select("count(id)").Row()
		row.Scan(&total)
		statics[res[i].ID] = total
	}

	// 汇总用户组存储策略
	policies := make(map[uint]model.Policy)
	for i := 0; i < len(res); i++ {
		for _, p := range res[i].PolicyList {
			if _, ok := policies[p]; !ok {
				policies[p], _ = model.GetPolicyByID(p)
			}
		}
	}

	return serializer.Response{Data: map[string]interface{}{
		"total":    total,
		"items":    res,
		"statics":  statics,
		"policies": policies,
	}}
}
