package controllers

import (
	"encoding/json"
	"fmt"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/authn"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/thumb"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/cloudreve/Cloudreve/v3/service/user"
	"github.com/duo-labs/webauthn/webauthn"
	"github.com/gin-gonic/gin"
)

// StartLoginAuthn 开始注册WebAuthn登录
func StartLoginAuthn(c *gin.Context) {
	userName := c.Param("username")
	expectedUser, err := model.GetActiveUserByEmail(userName)
	if err != nil {
		c.JSON(200, serializer.Err(serializer.CodeNotFound, "用户不存在", err))
		return
	}

	instance, err := authn.NewAuthnInstance()
	if err != nil {
		c.JSON(200, serializer.Err(serializer.CodeInternalSetting, "无法初始化Authn", err))
		return
	}

	options, sessionData, err := instance.BeginLogin(expectedUser)

	if err != nil {
		c.JSON(200, ErrorResponse(err))
		return
	}

	val, err := json.Marshal(sessionData)
	if err != nil {
		c.JSON(200, ErrorResponse(err))
		return
	}

	util.SetSession(c, map[string]interface{}{
		"registration-session": val,
	})
	c.JSON(200, serializer.Response{Code: 0, Data: options})
}

// FinishLoginAuthn 完成注册WebAuthn登录
func FinishLoginAuthn(c *gin.Context) {
	userName := c.Param("username")
	expectedUser, err := model.GetActiveUserByEmail(userName)
	if err != nil {
		c.JSON(200, serializer.Err(serializer.CodeCredentialInvalid, "用户邮箱或密码错误", err))
		return
	}

	sessionDataJSON := util.GetSession(c, "registration-session").([]byte)

	var sessionData webauthn.SessionData
	err = json.Unmarshal(sessionDataJSON, &sessionData)

	instance, err := authn.NewAuthnInstance()
	if err != nil {
		c.JSON(200, serializer.Err(serializer.CodeInternalSetting, "无法初始化Authn", err))
		return
	}

	_, err = instance.FinishLogin(expectedUser, sessionData, c.Request)

	if err != nil {
		c.JSON(200, serializer.Err(serializer.CodeCredentialInvalid, "登录验证失败", err))
		return
	}

	util.SetSession(c, map[string]interface{}{
		"user_id": expectedUser.ID,
	})
	c.JSON(200, serializer.BuildUserResponse(expectedUser))
}

// StartRegAuthn 开始注册WebAuthn信息
func StartRegAuthn(c *gin.Context) {
	currUser := CurrentUser(c)

	instance, err := authn.NewAuthnInstance()
	if err != nil {
		c.JSON(200, serializer.Err(serializer.CodeInternalSetting, "无法初始化Authn", err))
		return
	}

	options, sessionData, err := instance.BeginRegistration(currUser)

	if err != nil {
		c.JSON(200, ErrorResponse(err))
		return
	}

	val, err := json.Marshal(sessionData)
	if err != nil {
		c.JSON(200, ErrorResponse(err))
		return
	}

	util.SetSession(c, map[string]interface{}{
		"registration-session": val,
	})
	c.JSON(200, serializer.Response{Code: 0, Data: options})
}

// FinishRegAuthn 完成注册WebAuthn信息
func FinishRegAuthn(c *gin.Context) {
	currUser := CurrentUser(c)
	sessionDataJSON := util.GetSession(c, "registration-session").([]byte)

	var sessionData webauthn.SessionData
	err := json.Unmarshal(sessionDataJSON, &sessionData)

	instance, err := authn.NewAuthnInstance()
	if err != nil {
		c.JSON(200, serializer.Err(serializer.CodeInternalSetting, "无法初始化Authn", err))
		return
	}

	credential, err := instance.FinishRegistration(currUser, sessionData, c.Request)

	if err != nil {
		c.JSON(200, ErrorResponse(err))
		return
	}

	err = currUser.RegisterAuthn(credential)
	if err != nil {
		c.JSON(200, ErrorResponse(err))
		return
	}

	c.JSON(200, serializer.Response{
		Code: 0,
		Data: map[string]interface{}{
			"id":          credential.ID,
			"fingerprint": fmt.Sprintf("% X", credential.Authenticator.AAGUID),
		},
	})
}

// UserLogin 用户登录
func UserLogin(c *gin.Context) {
	var service user.UserLoginService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Login(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// UserRegister 用户注册
func UserRegister(c *gin.Context) {
	var service user.UserRegisterService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Register(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// User2FALogin 用户二步验证登录
func User2FALogin(c *gin.Context) {
	var service user.Enable2FA
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Login(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// UserSendReset 发送密码重设邮件
func UserSendReset(c *gin.Context) {
	var service user.UserResetEmailService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Reset(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// UserReset 重设密码
func UserReset(c *gin.Context) {
	var service user.UserResetService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Reset(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// UserActivate 用户激活
func UserActivate(c *gin.Context) {
	var service user.SettingService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.Activate(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// UserSignOut 用户退出登录
func UserSignOut(c *gin.Context) {
	util.DeleteSession(c, "user_id")
	c.JSON(200, serializer.Response{})
}

// UserMe 获取当前登录的用户
func UserMe(c *gin.Context) {
	currUser := CurrentUser(c)
	res := serializer.BuildUserResponse(*currUser)
	c.JSON(200, res)
}

// UserStorage 获取用户的存储信息
func UserStorage(c *gin.Context) {
	currUser := CurrentUser(c)
	res := serializer.BuildUserStorageResponse(*currUser)
	c.JSON(200, res)
}

// UserTasks 获取任务队列
func UserTasks(c *gin.Context) {
	var service user.SettingListService
	if err := c.ShouldBindQuery(&service); err == nil {
		res := service.ListTasks(c, CurrentUser(c))
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// UserSetting 获取用户设定
func UserSetting(c *gin.Context) {
	var service user.SettingService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.Settings(c, CurrentUser(c))
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// UseGravatar 设定头像使用全球通用
func UseGravatar(c *gin.Context) {
	u := CurrentUser(c)
	if err := u.Update(map[string]interface{}{"avatar": "gravatar"}); err != nil {
		c.JSON(200, serializer.Err(serializer.CodeDBError, "无法更新头像", err))
		return
	}
	c.JSON(200, serializer.Response{})
}

// UploadAvatar 从文件上传头像
func UploadAvatar(c *gin.Context) {
	// 取得头像上传大小限制
	maxSize := model.GetIntSetting("avatar_size", 2097152)
	if c.Request.ContentLength == -1 || c.Request.ContentLength > int64(maxSize) {
		request.BlackHole(c.Request.Body)
		c.JSON(200, serializer.Err(serializer.CodeUploadFailed, "头像尺寸太大", nil))
		return
	}

	// 取得上传的文件
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(200, serializer.Err(serializer.CodeIOFailed, "无法读取头像数据", err))
		return
	}

	// 初始化头像
	r, err := file.Open()
	if err != nil {
		c.JSON(200, serializer.Err(serializer.CodeIOFailed, "无法读取头像数据", err))
		return
	}
	avatar, err := thumb.NewThumbFromFile(r, file.Filename)
	if err != nil {
		c.JSON(200, serializer.Err(serializer.CodeIOFailed, "无法解析图像数据", err))
		return
	}

	// 创建头像
	u := CurrentUser(c)
	err = avatar.CreateAvatar(u.ID)
	if err != nil {
		c.JSON(200, serializer.Err(serializer.CodeIOFailed, "无法创建头像", err))
		return
	}

	// 保存头像标记
	if err := u.Update(map[string]interface{}{
		"avatar": "file",
	}); err != nil {
		c.JSON(200, serializer.Err(serializer.CodeDBError, "无法更新头像", err))
		return
	}

	c.JSON(200, serializer.Response{})
}

// GetUserAvatar 获取用户头像
func GetUserAvatar(c *gin.Context) {
	var service user.AvatarService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.Get(c)
		if res.Code == -301 {
			// 重定向到gravatar
			c.Redirect(301, res.Data.(string))
		}
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// UpdateOption 更改用户设定
func UpdateOption(c *gin.Context) {
	var service user.SettingUpdateService
	if err := c.ShouldBindUri(&service); err == nil {
		var (
			subService user.OptionsChangeHandler
			subErr     error
		)

		switch service.Option {
		case "nick":
			subService = &user.ChangerNick{}
		case "homepage":
			subService = &user.HomePage{}
		case "password":
			subService = &user.PasswordChange{}
		case "2fa":
			subService = &user.Enable2FA{}
		case "authn":
			subService = &user.DeleteWebAuthn{}
		case "theme":
			subService = &user.ThemeChose{}
		default:
			subService = &user.ChangerNick{}
		}

		subErr = c.ShouldBindJSON(subService)
		if subErr != nil {
			c.JSON(200, ErrorResponse(subErr))
			return
		}

		res := subService.Update(c, CurrentUser(c))
		c.JSON(200, res)

	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// UserInit2FA 初始化二步验证
func UserInit2FA(c *gin.Context) {
	var service user.SettingService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.Init2FA(c, CurrentUser(c))
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}
