package user

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/auth"
	"github.com/HFO4/cloudreve/pkg/email"
	"github.com/HFO4/cloudreve/pkg/hashid"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	"net/url"
	"strings"
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
	if isCaptchaRequired {
		captchaID := util.GetSession(c, "captchaID")
		util.DeleteSession(c, "captchaID")
		if captchaID == nil || !base64Captcha.VerifyCaptcha(captchaID.(string), service.CaptchaCode) {
			return serializer.ParamErr("验证码错误", nil)
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
			strings.ReplaceAll(finalURL.String(), "/activate", "/#/activate"),
		)
		if err := email.Send(user.Email, title, body); err != nil {
			return serializer.Err(serializer.CodeInternalSetting, "无法发送激活邮件", err)
		}

		return serializer.Response{Code: 203}
	}

	return serializer.Response{}
}
