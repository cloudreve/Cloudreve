package vas

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/qq"
	"github.com/HFO4/cloudreve/pkg/request"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/pkg/thumb"
	"github.com/HFO4/cloudreve/pkg/util"
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
		return serializer.Err(serializer.CodeSignExpired, "请求过期，请重试", nil)
	}
	util.DeleteSession(c, "qq_login_secret")

	// 获取OpenID
	credential, err := qq.Callback(service.Code)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, "无法获取登录状态", err)
	}

	// 如果已登录，则绑定已有用户
	if user != nil {

		if user.OpenID != "" {
			return serializer.Err(serializer.CodeCallbackError, "您已绑定了QQ账号，请先解除绑定", nil)
		}
		if err := user.Update(map[string]interface{}{"open_id": credential.OpenID}); err != nil {
			return serializer.DBErr("绑定失败", err)
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
		return serializer.Err(serializer.CodeNoPermissionErr, "此QQ号未绑定任何账号", nil)
	}

	// 获取用户信息
	userInfo, err := qq.GetUserInfo(credential)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, "无法获取用户信息", err)
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
		return serializer.DBErr("此邮箱已被使用", err)
	}

	// 下载头像
	r := request.HTTPClient{}
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
