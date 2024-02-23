package vas

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/qq"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/thumb"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-gonic/gin"
)

// QQCallbackService QQ互联回调处理服务
type QQCallbackService struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state" binding:"required"`
}

// Callback 处理QQ互联回调
func (service *QQCallbackService) Callback(c *gin.Context, user *model.User) serializer.Response {

	state := util.GetSession(c, "qq_login_secret")
	if stateStr, ok := state.(string); !ok || stateStr != service.State {
		return serializer.Err(serializer.CodeSignExpired, "", nil)
	}
	util.DeleteSession(c, "qq_login_secret")

	// 获取OpenID
	credential, err := qq.Callback(service.Code)
	if err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "Failed to get session status", err)
	}

	// 如果已登录，则绑定已有用户
	if user != nil {

		if user.OpenID != "" {
			return serializer.Err(serializer.CodeQQBindConflict, "", nil)
		}

		// OpenID 是否重复
		if _, err := model.GetActiveUserByOpenID(credential.OpenID); err == nil {
			return serializer.Err(serializer.CodeQQBindOtherAccount, "", nil)
		}

		if err := user.Update(map[string]interface{}{"open_id": credential.OpenID}); err != nil {
			return serializer.DBErr("Failed to update user open id", err)
		}
		return serializer.Response{
			Data: "/setting",
		}

	}

	// 未登录，尝试查找用户
	if expectedUser, err := model.GetActiveUserByOpenID(credential.OpenID); err == nil {
		// 用户绑定了此QQ，设定为登录状态
		util.SetSession(c, map[string]interface{}{
			"user_id": expectedUser.ID,
		})
		res := serializer.BuildUserResponse(expectedUser)
		res.Code = 203
		return res

	}

	// 无匹配用户，创建新用户
	if !model.IsTrueVal(model.GetSettingByName("qq_direct_login")) {
		return serializer.Err(serializer.CodeQQNotLinked, "", nil)
	}

	// 获取用户信息
	userInfo, err := qq.GetUserInfo(credential)
	if err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "Failed to fetch user info", err)
	}

	// 生成邮箱地址
	fakeEmail := util.RandStringRunes(16) + "@login.qq.com"

	// 创建用户
	defaultGroup := model.GetIntSetting("default_group", 2)

	newUser := model.NewUser()
	newUser.Email = fakeEmail
	newUser.Nick = userInfo.Nick
	newUser.SetPassword("")
	newUser.Status = model.Active
	newUser.GroupID = uint(defaultGroup)
	newUser.OpenID = credential.OpenID
	newUser.Avatar = "file"

	// 创建用户
	if err := model.DB.Create(&newUser).Error; err != nil {
		return serializer.Err(serializer.CodeEmailExisted, "", err)
	}

	// 下载头像
	r := request.NewClient()
	rawAvatar := r.Request("GET", userInfo.Avatar, nil)
	if avatar, err := thumb.NewThumbFromFile(rawAvatar.Response.Body, "avatar.jpg"); err == nil {
		avatar.CreateAvatar(newUser.ID)
	}

	// 登录
	util.SetSession(c, map[string]interface{}{"user_id": newUser.ID})

	newUser, _ = model.GetActiveUserByID(newUser.ID)

	res := serializer.BuildUserResponse(newUser)
	res.Code = 203
	return res
}
