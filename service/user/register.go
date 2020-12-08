package user

import (
	"net/url"
	"strings"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/email"
	"github.com/cloudreve/Cloudreve/v3/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v3/pkg/recaptcha"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
)

// UserRegisterService 管理用户注册的服务
type UserRegisterService struct {
	//TODO 细致调整验证规则
	UserName    string `form:"userName" json:"userName" binding:"required,email"`
	Password    string `form:"Password" json:"Password" binding:"required,min=4,max=64"`
	CaptchaCode string `form:"captchaCode" json:"captchaCode"`
}

// Register 新用户注册
func (service *UserRegisterService) Register(c *gin.Context) serializer.Response {
	// 相关设定
	options := model.GetSettingByNames("email_active", "reg_captcha")
	// 检查验证码
	isCaptchaRequired := model.IsTrueVal(options["reg_captcha"])
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

	// 相关设定
	isEmailRequired := model.IsTrueVal(options["email_active"])
	defaultGroup := model.GetIntSetting("default_group", 2)

	// 创建新的用户对象
	user := model.NewUser()
	user.Email = service.UserName
	user.Nick = strings.Split(service.UserName, "@")[0]
	user.SetPassword(service.Password)
	user.Status = model.Active
	if isEmailRequired {
		user.Status = model.NotActivicated
	}
	user.GroupID = uint(defaultGroup)

	// 创建用户
	if err := model.DB.Create(&user).Error; err != nil {
		return serializer.DBErr("此邮箱已被使用", err)
	}

	// 发送激活邮件
	if isEmailRequired {

		// 签名激活请求API
		base := model.GetSiteURL()
		userID := hashid.HashID(user.ID, hashid.UserID)
		controller, _ := url.Parse("/api/v3/user/activate/" + userID)
		activateURL, err := auth.SignURI(auth.General, base.ResolveReference(controller).String(), 86400)
		if err != nil {
			return serializer.Err(serializer.CodeEncryptError, "无法签名激活URL", err)
		}

		// 取得签名
		credential := activateURL.Query().Get("sign")

		// 生成对用户访问的激活地址
		controller, _ = url.Parse("/activate")
		finalURL := base.ResolveReference(controller)
		queries := finalURL.Query()
		queries.Add("id", userID)
		queries.Add("sign", credential)
		finalURL.RawQuery = queries.Encode()

		// 返送激活邮件
		title, body := email.NewActivationEmail(user.Email,
			finalURL.String(),
		)
		if err := email.Send(user.Email, title, body); err != nil {
			return serializer.Err(serializer.CodeInternalSetting, "无法发送激活邮件", err)
		}

		return serializer.Response{Code: 203}
	}

	return serializer.Response{}
}

// Activate 激活用户
func (service *SettingService) Activate(c *gin.Context) serializer.Response {
	// 查找待激活用户
	uid, _ := c.Get("object_id")
	user, err := model.GetUserByID(uid.(uint))
	if err != nil {
		return serializer.Err(serializer.CodeNotFound, "用户不存在", err)
	}

	// 检查状态
	if user.Status != model.NotActivicated {
		return serializer.Err(serializer.CodeNoPermissionErr, "该用户无法被激活", nil)
	}

	// 激活用户
	user.SetStatus(model.Active)

	return serializer.Response{Data: user.Email}
}
