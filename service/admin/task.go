package admin

import (
	"strings"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/task"
	"github.com/gin-gonic/gin"
)

// TaskBatchService 任务批量操作服务
type TaskBatchService struct {
	ID []uint `json:"id" binding:"min=1"`
}

// ImportTaskService 导入任务
type ImportTaskService struct {
	UID       uint   `json:"uid" binding:"required"`
	PolicyID  uint   `json:"policy_id" binding:"required"`
	Src       string `json:"src" binding:"required,min=1,max=65535"`
	Dst       string `json:"dst" binding:"required,min=1,max=65535"`
	Recursive bool   `json:"recursive"`
}

// Create 新建导入任务
func (service *ImportTaskService) Create(c *gin.Context, user *model.User) serializer.Response {
	// 创建任务
	job, err := task.NewImportTask(service.UID, service.PolicyID, service.Src, service.Dst, service.Recursive)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, "任务创建失败", err)
	}
	task.TaskPoll.Submit(job)
	return serializer.Response{}
}

// Delete 删除任务
func (service *TaskBatchService) Delete(c *gin.Context) serializer.Response {
	if err := model.DB.Where("id in (?)", service.ID).Delete(&model.Download{}).Error; err != nil {
		return serializer.DBErr("无法删除任务", err)
	}
	return serializer.Response{}
}

// DeleteGeneral 删除常规任务
func (service *TaskBatchService) DeleteGeneral(c *gin.Context) serializer.Response {
	if err := model.DB.Where("id in (?)", service.ID).Delete(&model.Task{}).Error; err != nil {
		return serializer.DBErr("无法删除任务", err)
	}
	return serializer.Response{}
}

// Tasks 列出常规任务
func (service *AdminListService) Tasks() serializer.Response {
	var res []model.Task
	total := 0

	tx := model.DB.Model(&model.Task{})
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

	// 查询对应用户，同时计算HashID
	users := make(map[uint]model.User)
	for _, file := range res {
		users[file.UserID] = model.User{}
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
	}}
}

// Downloads 列出离线下载任务
func (service *AdminListService) Downloads() serializer.Response {
	var res []model.Download
	total := 0

	tx := model.DB.Model(&model.Download{})
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

	// 查询对应用户，同时计算HashID
	users := make(map[uint]model.User)
	for _, file := range res {
		users[file.UserID] = model.User{}
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
	}}
}
