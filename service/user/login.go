package user

import (
	"fmt"
	"net/url"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/email"
	"github.com/cloudreve/Cloudreve/v3/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v3/pkg/recaptcha"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	"github.com/pquerna/otp/totp"
)

// UserLoginService 管理用户登录的服务
type UserLoginService struct {
	//TODO 细致调整验证规则
	UserName    string `form:"userName" json:"userName" binding:"required,email"`
	Password    string `form:"Password" json:"Password" binding:"required,min=4,max=64"`
	CaptchaCode string `form:"captchaCode" json:"captchaCode"`
}

// UserResetEmailService 发送密码重设邮件服务
type UserResetEmailService struct {
	UserName    string `form:"userName" json:"userName" binding:"required,email"`
	CaptchaCode string `form:"captchaCode" json:"captchaCode"`
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
		return serializer.Err(serializer.CodeNotFound, "重设链接无效", err)
	}

	// 检查重设会话
	resetSession, exist := cache.Get(fmt.Sprintf("user_reset_%d", uid))
	if !exist || resetSession.(string) != service.Secret {
		return serializer.Err(serializer.CodeNotFound, "链接已过期", err)
	}

	// 重设用户密码
	user, err := model.GetActiveUserByID(uid)
	if err != nil {
		return serializer.Err(serializer.CodeNotFound, "用户不存在", err)
	}

	user.SetPassword(service.Password)
	if err := user.Update(map[string]interface{}{"password": user.Password}); err != nil {
		return serializer.DBErr("无法重设密码", err)
	}

	cache.Deletes([]string{fmt.Sprintf("%d", uid)}, "user_reset_")
	return serializer.Response{}
}

// Reset 发送密码重设邮件
func (service *UserResetEmailService) Reset(c *gin.Context) serializer.Response {
	// 检查验证码
	isCaptchaRequired := model.IsTrueVal(model.GetSettingByName("forget_captcha"))
	useRecaptcha := model.IsTrueVal(model.GetSettingByName("captcha_IsUseReCaptcha"))
	recaptchaSecret := model.GetSettingByName("captcha_ReCaptchaSecret")
	if isCaptchaRequired && !useRecaptcha {
		captchaID := util.GetSession(c, "captchaID")
		util.DeleteSession(c, "captchaID")
		if captchaID == nil || !base64Captcha.VerifyCaptcha(captchaID.(string), service.CaptchaCode) {
			return serializer.ParamErr("验证码错误", nil)
		}
	} else if isCaptchaRequired && useRecaptcha {
		captcha, err := recaptcha.NewReCAPTCHA(recaptchaSecret, recaptcha.V2, 10*time.Second)
		if err != nil {
			util.Log().Error(err.Error())
		}
		err = captcha.Verify(service.CaptchaCode)
		if err != nil {
			util.Log().Error(err.Error())
			return serializer.ParamErr("验证失败，请刷新网页后再次验证", nil)
		}
	}

	// 查找用户
	if user, err := model.GetUserByEmail(service.UserName); err == nil {

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
			return serializer.Err(serializer.CodeInternalSetting, "无法发送密码重设邮件", err)
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
			return serializer.Err(serializer.CodeNotFound, "用户不存在", nil)
		}

		// 验证二步验证代码
		if !totp.Validate(service.Code, expectedUser.TwoFactor) {
			return serializer.ParamErr("验证代码不正确", nil)
		}

		//登陆成功，清空并设置session
		util.DeleteSession(c, "2fa_user_id")
		util.SetSession(c, map[string]interface{}{
			"user_id": expectedUser.ID,
		})

		return serializer.BuildUserResponse(expectedUser)
	}

	return serializer.Err(serializer.CodeNotFound, "登录会话不存在", nil)
}

// Login 用户登录函数
func (service *UserLoginService) Login(c *gin.Context) serializer.Response {
	isCaptchaRequired := model.GetSettingByName("login_captcha")
	useRecaptcha := model.GetSettingByName("captcha_IsUseReCaptcha")
	recaptchaSecret := model.GetSettingByName("captcha_ReCaptchaSecret")
	expectedUser, err := model.GetUserByEmail(service.UserName)

	if (model.IsTrueVal(isCaptchaRequired)) && !(model.IsTrueVal(useRecaptcha)) {
		// TODO 验证码校验
		captchaID := util.GetSession(c, "captchaID")
		util.DeleteSession(c, "captchaID")
		if captchaID == nil || !base64Captcha.VerifyCaptcha(captchaID.(string), service.CaptchaCode) {
			return serializer.ParamErr("验证码错误", nil)
		}
	} else if (model.IsTrueVal(isCaptchaRequired)) && (model.IsTrueVal(useRecaptcha)) {
		captcha, err := recaptcha.NewReCAPTCHA(recaptchaSecret, recaptcha.V2, 10*time.Second)
		if err != nil {
			util.Log().Error(err.Error())
		}
		err = captcha.Verify(service.CaptchaCode)
		if err != nil {
			util.Log().Error(err.Error())
			return serializer.ParamErr("验证失败，请刷新网页后再次验证", nil)
		}
	}

	// 一系列校验
	if err != nil {
		return serializer.Err(serializer.CodeCredentialInvalid, "用户邮箱或密码错误", err)
	}
	if authOK, _ := expectedUser.CheckPassword(service.Password); !authOK {
		return serializer.Err(serializer.CodeCredentialInvalid, "用户邮箱或密码错误", nil)
	}
	if expectedUser.Status == model.Baned || expectedUser.Status == model.OveruseBaned {
		return serializer.Err(403, "该账号已被封禁", nil)
	}
	if expectedUser.Status == model.NotActivicated {
		return serializer.Err(403, "该账号未激活", nil)
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
