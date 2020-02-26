package admin

import (
	"fmt"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/pkg/util"
	"os"
	"path/filepath"
)

// PathTestService 本地路径测试服务
type PathTestService struct {
	Path string `json:"path" binding:"required"`
}

// AddPolicyService 存储策略添加服务
type AddPolicyService struct {
	Policy model.Policy `json:"policy" binding:"required"`
}

// Add 添加存储策略
func (service *AddPolicyService) Add() serializer.Response {
	if err := model.DB.Create(&service.Policy).Error; err != nil {
		return serializer.ParamErr("存储策略添加失败", err)
	}
	return serializer.Response{}
}

// Test 测试本地路径
func (service *PathTestService) Test() serializer.Response {
	policy := model.Policy{DirNameRule: service.Path}
	path := policy.GeneratePath(1, "/My File")
	path = filepath.Join(path, "test.txt")
	file, err := util.CreatNestedFile(path)
	if err != nil {
		return serializer.ParamErr(fmt.Sprintf("无法创建路径 %s , %s", path, err.Error()), nil)
	}

	file.Close()
	os.Remove(path)

	return serializer.Response{}
}

// Policies 列出存储策略
func (service *AdminListService) Policies() serializer.Response {
	var res []model.Policy
	total := 0

	tx := model.DB.Model(&model.Policy{})
	if service.OrderBy != "" {
		tx = tx.Order(service.OrderBy)
	}

	for k, v := range service.Conditions {
		tx = tx.Where("? = ?", k, v)
	}

	// 计算总数用于分页
	tx.Count(&total)

	// 查询记录
	tx.Limit(service.PageSize).Offset((service.Page - 1) * service.PageSize).Find(&res)

	// 统计每个策略的文件使用
	statics := make(map[uint][2]int, len(res))
	for i := 0; i < len(res); i++ {
		total := [2]int{}
		row := model.DB.Model(&model.File{}).Where("policy_id = ?", res[i].ID).
			Select("count(id),sum(size)").Row()
		row.Scan(&total[0], &total[1])
		statics[res[i].ID] = total
	}

	return serializer.Response{Data: map[string]interface{}{
		"total":   total,
		"items":   res,
		"statics": statics,
	}}
}
