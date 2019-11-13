package user

import (
	"cloudreve/models"
	"cloudreve/pkg/serializer"
	"cloudreve/pkg/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
)

// UserLoginService 管理用户登录的服务
type UserLoginService struct {
	//TODO 细致调整验证规则
	UserName    string `form:"userName" json:"userName" binding:"required,email"`
	Password    string `form:"Password" json:"Password" binding:"required,min=4,max=64"`
	CaptchaCode string `form:"captchaCode" json:"captchaCode"`
}

// Login 用户登录函数
func (service *UserLoginService) Login(c *gin.Context) serializer.Response {
	isCaptchaRequired := model.GetSettingByName("login_captcha")
	expectedUser, err := model.GetUserByEmail(service.UserName)

	if model.IsTrueVal(isCaptchaRequired) {
		// TODO 验证码校验
		captchaID := util.GetSession(c, "captchaID")
		if captchaID == nil || !base64Captcha.VerifyCaptcha(captchaID.(string), service.CaptchaCode) {
			util.DeleteSession(c, "captchaID")
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
	if expectedUser.Status == model.Baned {
		return serializer.Err(403, "该账号已被封禁", nil)
	}
	if expectedUser.Status == model.NotActivicated {
		return serializer.Err(403, "该账号未激活", nil)
	}

	if expectedUser.TwoFactor != "" {
		//TODO 二步验证处理
	}

	//登陆成功，清空并设置session
	util.SetSession(c, map[string]interface{}{
		"user_id": expectedUser.ID,
	})

	fmt.Println(expectedUser)

	return serializer.BuildUserResponse(expectedUser)

}
