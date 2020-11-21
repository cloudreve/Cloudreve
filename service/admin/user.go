package admin

import (
	"context"
	"strings"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
)

// AddUserService 用户添加服务
type AddUserService struct {
	User     model.User `json:"User" binding:"required"`
	Password string     `json:"password"`
}

// UserService 用户ID服务
type UserService struct {
	ID uint `uri:"id" json:"id" binding:"required"`
}

// UserBatchService 用户批量操作服务
type UserBatchService struct {
	ID []uint `json:"id" binding:"min=1"`
}

// Ban 封禁/解封用户
func (service *UserService) Ban() serializer.Response {
	user, err := model.GetUserByID(service.ID)
	if err != nil {
		return serializer.Err(serializer.CodeNotFound, "用户不存在", err)
	}

	if user.ID == 1 {
		return serializer.Err(serializer.CodeNoPermissionErr, "无法封禁初始用户", err)
	}

	if user.Status == model.Active {
		user.SetStatus(model.Baned)
	} else {
		user.SetStatus(model.Active)
	}

	return serializer.Response{Data: user.Status}
}

// Delete 删除用户
func (service *UserBatchService) Delete() serializer.Response {
	for _, uid := range service.ID {
		user, err := model.GetUserByID(uid)
		if err != nil {
			return serializer.Err(serializer.CodeNotFound, "用户不存在", err)
		}

		// 不能删除初始用户
		if uid == 1 {
			return serializer.Err(serializer.CodeNoPermissionErr, "无法删除初始用户", err)
		}

		// 删除与此用户相关的所有资源

		fs, err := filesystem.NewFileSystem(&user)
		// 删除所有文件
		root, err := fs.User.Root()
		if err != nil {
			return serializer.Err(serializer.CodeNotFound, "无法找到用户根目录", err)
		}
		fs.Delete(context.Background(), []uint{root.ID}, []uint{}, false)

		// 删除相关任务
		model.DB.Where("user_id = ?", uid).Delete(&model.Download{})
		model.DB.Where("user_id = ?", uid).Delete(&model.Task{})

		// 删除标签
		model.DB.Where("user_id = ?", uid).Delete(&model.Tag{})

		// 删除WebDAV账号
		model.DB.Where("user_id = ?", uid).Delete(&model.Webdav{})

		// 删除此用户
		model.DB.Unscoped().Delete(user)

	}
	return serializer.Response{}
}

// Get 获取用户详情
func (service *UserService) Get() serializer.Response {
	group, err := model.GetUserByID(service.ID)
	if err != nil {
		return serializer.Err(serializer.CodeNotFound, "用户不存在", err)
	}

	return serializer.Response{Data: group}
}

// Add 添加用户
func (service *AddUserService) Add() serializer.Response {
	if service.User.ID > 0 {

		user, _ := model.GetUserByID(service.User.ID)
		if service.Password != "" {
			user.SetPassword(service.Password)
		}

		// 只更新必要字段
		user.Nick = service.User.Nick
		user.Email = service.User.Email
		user.GroupID = service.User.GroupID
		user.Status = service.User.Status

		// 检查愚蠢操作
		if user.ID == 1 && user.GroupID != 1 {
			return serializer.ParamErr("无法更改初始用户的用户组", nil)
		}

		if err := model.DB.Save(&user).Error; err != nil {
			return serializer.ParamErr("用户保存失败", err)
		}
	} else {
		service.User.SetPassword(service.Password)
		if err := model.DB.Create(&service.User).Error; err != nil {
			return serializer.ParamErr("用户组添加失败", err)
		}
	}

	return serializer.Response{Data: service.User.ID}
}

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

	if len(service.Searches) > 0 {
		search := ""
		for k, v := range service.Searches {
			search += (k + " like '%" + v + "%' OR ")
		}
		search = strings.TrimSuffix(search, " OR ")
		tx = tx.Where(search)
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
