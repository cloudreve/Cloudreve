package admin

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/gofrs/uuid"
)

// GenerateRedeemsService 兑换码生成服务
type GenerateRedeemsService struct {
	Num  int   `json:"num" binding:"required,min=1,max=100"`
	ID   int64 `json:"id"`
	Time int   `json:"time" binding:"required,min=1"`
	Type int   `json:"type" binding:"min=0,max=2"`
}

// SingleIDService 单ID服务
type SingleIDService struct {
	ID uint `uri:"id" binding:"required"`
}

// DeleteRedeem 删除兑换码
func (service *SingleIDService) DeleteRedeem() serializer.Response {
	if err := model.DB.Where("id = ?", service.ID).Delete(&model.Redeem{}).Error; err != nil {
		return serializer.DBErr("无法删除兑换码", err)
	}

	return serializer.Response{}
}

// Generate 生成兑换码
func (service *GenerateRedeemsService) Generate() serializer.Response {
	res := make([]string, service.Num)
	redeem := model.Redeem{}

	// 开始事务
	tx := model.DB.Begin()
	if err := tx.Error; err != nil {
		return serializer.DBErr("无法开启事务", err)
	}

	// 创建每个兑换码
	for i := 0; i < service.Num; i++ {
		redeem.Model.ID = 0
		redeem.Num = service.Time
		redeem.Type = service.Type
		redeem.ProductID = service.ID
		redeem.Used = false

		// 生成唯一兑换码
		u2, err := uuid.NewV4()
		if err != nil {
			tx.Rollback()
			return serializer.Err(serializer.CodeInternalSetting, "无法生成兑换码", err)
		}

		redeem.Code = u2.String()
		if err := tx.Create(&redeem).Error; err != nil {
			tx.Rollback()
			return serializer.DBErr("无法创建兑换码记录", err)
		}

		res[i] = redeem.Code
	}

	if err := tx.Commit().Error; err != nil {
		return serializer.DBErr("无法创建兑换码记录", err)
	}

	return serializer.Response{Data: res}

}

// Redeems 列出激活码
func (service *AdminListService) Redeems() serializer.Response {
	var res []model.Redeem
	total := 0

	tx := model.DB.Model(&model.Redeem{})
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

	return serializer.Response{Data: map[string]interface{}{
		"total": total,
		"items": res,
	}}
}
