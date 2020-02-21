package user

import (
	"github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/pkg/util"
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
	expectedUser, err := model.GetUserByEmail(service.UserName)

	if model.IsTrueVal(isCaptchaRequired) {
		// TODO 验证码校验
		captchaID := util.GetSession(c, "captchaID")
		if captchaID == nil || !base64Captcha.VerifyCaptcha(captchaID.(string), service.CaptchaCode) {
			return serializer.ParamErr("验证码错误", nil)
		}
	}

	// 一系列校验
	if err != nil {
		return serializer.Err(401, "用户邮箱或密码错误", err)
	}
	if authOK, _ := expectedUser.CheckPassword(service.Password); !authOK {
		return serializer.Err(401, "用户邮箱或密码错误", nil)
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
