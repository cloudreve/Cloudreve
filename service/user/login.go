package user

import (
	"fmt"
	"net/url"

	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/gofrs/uuid"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/email"
	"github.com/cloudreve/Cloudreve/v3/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
)

// UserLoginService 管理用户登录的服务
type UserLoginService struct {
	//TODO 细致调整验证规则
	UserName string `form:"userName" json:"userName" binding:"required,email"`
	Password string `form:"Password" json:"Password" binding:"required,min=4,max=64"`
}

// UserResetEmailService 发送密码重设邮件服务
type UserResetEmailService struct {
	UserName string `form:"userName" json:"userName" binding:"required,email"`
}

// UserResetService 密码重设服务
type UserResetService struct {
	Password string `form:"Password" json:"Password" binding:"required,min=4,max=64"`
	ID       string `json:"id" binding:"required"`
	Secret   string `json:"secret" binding:"required"`
}

// Reset 重设密码
func (service *UserResetService) Reset(c *gin.Context) serializer.Response {
	// 取得原始用户ID
	uid, err := hashid.DecodeHashID(service.ID, hashid.UserID)
	if err != nil {
		return serializer.Err(serializer.CodeInvalidTempLink, "Invalid link", err)
	}

	// 检查重设会话
	resetSession, exist := cache.Get(fmt.Sprintf("user_reset_%d", uid))
	if !exist || resetSession.(string) != service.Secret {
		return serializer.Err(serializer.CodeTempLinkExpired, "Link is expired", err)
	}

	// 重设用户密码
	user, err := model.GetActiveUserByID(uid)
	if err != nil {
		return serializer.Err(serializer.CodeUserNotFound, "User not found", nil)
	}

	user.SetPassword(service.Password)
	if err := user.Update(map[string]interface{}{"password": user.Password}); err != nil {
		return serializer.DBErr("Failed to reset password", err)
	}

	cache.Deletes([]string{fmt.Sprintf("%d", uid)}, "user_reset_")
	return serializer.Response{}
}

// Reset 发送密码重设邮件
func (service *UserResetEmailService) Reset(c *gin.Context) serializer.Response {
	// 查找用户
	if user, err := model.GetUserByEmail(service.UserName); err == nil {

		if user.Status == model.Baned || user.Status == model.OveruseBaned {
			return serializer.Err(serializer.CodeUserBaned, "This user is banned", nil)
		}
		if user.Status == model.NotActivicated {
			return serializer.Err(serializer.CodeUserNotActivated, "This user is not activated", nil)
		}
		// 创建密码重设会话
		secret := util.RandStringRunes(32)
		cache.Set(fmt.Sprintf("user_reset_%d", user.ID), secret, 3600)

		// 生成用户访问的重设链接
		controller, _ := url.Parse("/reset")
		finalURL := model.GetSiteURL().ResolveReference(controller)
		queries := finalURL.Query()
		queries.Add("id", hashid.HashID(user.ID, hashid.UserID))
		queries.Add("sign", secret)
		finalURL.RawQuery = queries.Encode()

		// 发送密码重设邮件
		title, body := email.NewResetEmail(user.Nick, finalURL.String())
		if err := email.Send(user.Email, title, body); err != nil {
			return serializer.Err(serializer.CodeFailedSendEmail, "Failed to send email", err)
		}

	}

	return serializer.Response{}
}

// Login 二步验证继续登录
func (service *Enable2FA) Login(c *gin.Context) serializer.Response {
	if uid, ok := util.GetSession(c, "2fa_user_id").(uint); ok {
		// 查找用户
		expectedUser, err := model.GetActiveUserByID(uid)
		if err != nil {
			return serializer.Err(serializer.CodeUserNotFound, "User not found", nil)
		}

		// 验证二步验证代码
		if !totp.Validate(service.Code, expectedUser.TwoFactor) {
			return serializer.Err(serializer.Code2FACodeErr, "2FA code not correct", nil)
		}

		//登陆成功，清空并设置session
		util.DeleteSession(c, "2fa_user_id")
		util.SetSession(c, map[string]interface{}{
			"user_id": expectedUser.ID,
		})

		return serializer.BuildUserResponse(expectedUser)
	}

	return serializer.Err(serializer.CodeLoginSessionNotExist, "Login session not exist", nil)
}

// Login 用户登录函数
func (service *UserLoginService) Login(c *gin.Context) serializer.Response {
	expectedUser, err := model.GetUserByEmail(service.UserName)
	// 一系列校验
	if err != nil {
		return serializer.Err(serializer.CodeCredentialInvalid, "Wrong password or email address", err)
	}
	if authOK, _ := expectedUser.CheckPassword(service.Password); !authOK {
		return serializer.Err(serializer.CodeCredentialInvalid, "Wrong password or email address", nil)
	}
	if expectedUser.Status == model.Baned || expectedUser.Status == model.OveruseBaned {
		return serializer.Err(serializer.CodeUserBaned, "This account has been blocked", nil)
	}
	if expectedUser.Status == model.NotActivicated {
		return serializer.Err(serializer.CodeUserNotActivated, "This account is not activated", nil)
	}

	if expectedUser.TwoFactor != "" {
		// 需要二步验证
		util.SetSession(c, map[string]interface{}{
			"2fa_user_id": expectedUser.ID,
		})
		return serializer.Response{Code: 203}
	}

	//登陆成功，清空并设置session
	util.SetSession(c, map[string]interface{}{
		"user_id": expectedUser.ID,
	})

	return serializer.BuildUserResponse(expectedUser)

}

// CopySessionService service for copy user session
type CopySessionService struct {
	ID string `uri:"id" binding:"required,uuid4"`
}

const CopySessionTTL = 60

// Prepare generates the URL with short expiration duration
func (s *CopySessionService) Prepare(c *gin.Context, user *model.User) serializer.Response {
	// 用户组有效期
	urlID := uuid.Must(uuid.NewV4())
	if err := cache.Set(fmt.Sprintf("copy_session_%s", urlID.String()), user.ID, CopySessionTTL); err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "Failed to create copy session", err)
	}

	base := model.GetSiteURL()
	apiBaseURI, _ := url.Parse("/api/v3/user/session/copy/" + urlID.String())
	apiURL := base.ResolveReference(apiBaseURI)
	res, err := auth.SignURI(auth.General, apiURL.String(), CopySessionTTL)
	if err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "Failed to sign temp URL", err)
	}

	return serializer.Response{
		Data: res.String(),
	}
}

// Copy a new session from active session, refresh max-age
func (s *CopySessionService) Copy(c *gin.Context) serializer.Response {
	// 用户组有效期
	cacheKey := fmt.Sprintf("copy_session_%s", s.ID)
	uid, ok := cache.Get(cacheKey)
	if !ok {
		return serializer.Err(serializer.CodeNotFound, "", nil)
	}

	cache.Deletes([]string{cacheKey}, "")
	util.SetSession(c, map[string]interface{}{
		"user_id": uid.(uint),
	})

	return serializer.Response{}
}
