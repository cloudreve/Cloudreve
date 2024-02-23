package admin

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
)

// ReportBatchService 任务批量操作服务
type ReportBatchService struct {
	ID []uint `json:"id" binding:"min=1"`
}

// Reports 批量删除举报
func (service *ReportBatchService) Delete() serializer.Response {
	if err := model.DB.Where("id in (?)", service.ID).Delete(&model.Report{}).Error; err != nil {
		return serializer.DBErr("Failed to change report status", err)
	}
	return serializer.Response{}
}

// Reports 列出待处理举报
func (service *AdminListService) Reports() serializer.Response {
	var res []model.Report
	total := 0

	tx := model.DB.Model(&model.Report{})
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

	// 计算分享的 HashID
	hashIDs := make(map[uint]string, len(res))
	for _, report := range res {
		hashIDs[report.Share.ID] = hashid.HashID(report.Share.ID, hashid.ShareID)
	}

	// 查询对应用户
	users := make(map[uint]model.User)
	for _, report := range res {
		users[report.Share.UserID] = model.User{}
	}

	userIDs := make([]uint, 0, len(users))
	for k := range users {
		userIDs = append(userIDs, k)
	}

	var userList []model.User
	model.DB.Where("id in (?)", userIDs).Find(&userList)

	for _, v := range userList {
		users[v.ID] = v
	}

	return serializer.Response{Data: map[string]interface{}{
		"total": total,
		"items": res,
		"users": users,
		"ids":   hashIDs,
	}}
}
