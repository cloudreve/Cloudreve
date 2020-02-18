package user

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/gin-gonic/gin"
)

// SettingService 通用设置服务
type SettingService struct {
}

// SettingListService 通用设置列表服务
type SettingListService struct {
	Page int `form:"page" binding:"required,min=1"`
}

// ListTasks 列出任务
func (service *SettingListService) ListTasks(c *gin.Context, user *model.User) serializer.Response {
	tasks, total := model.ListTasks(user.ID, service.Page, 10, "updated_at desc")
	return serializer.BuildTaskList(tasks, total)
}

// Policy 获取用户存储策略设置
func (service *SettingService) Policy(c *gin.Context, user *model.User) serializer.Response {
	// 取得用户可用存储策略
	available := make([]model.Policy, 0, len(user.Group.PolicyList))
	for _, id := range user.Group.PolicyList {
		if policy, err := model.GetPolicyByID(id); err == nil {
			available = append(available, policy)
		}
	}

	// 取得用户当前策略
	current := user.Policy

	return serializer.BuildPolicySettingRes(available, &current)
}
