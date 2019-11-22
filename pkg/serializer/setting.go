package serializer

import model "github.com/HFO4/cloudreve/models"

// SiteConfig 站点全局设置序列
type SiteConfig struct {
	SiteName      string `json:"title"`
	LoginCaptcha  bool   `json:"loginCaptcha"`
	RegCaptcha    bool   `json:"regCaptcha"`
	ForgetCaptcha bool   `json:"forgetCaptcha"`
	EmailActive   bool   `json:"emailActive"`
	QQLogin       bool   `json:"QQLogin"`
	Themes        string `json:"themes"`
	DefaultTheme  string `json:"defaultTheme"`
}

func checkSettingValue(setting map[string]string, key string) string {
	if v, ok := setting[key]; ok {
		return v
	}
	return ""
}

// BuildSiteConfig 站点全局设置
func BuildSiteConfig(settings map[string]string) Response {
	return Response{
		Data: SiteConfig{
			SiteName:      checkSettingValue(settings, "siteName"),
			LoginCaptcha:  model.IsTrueVal(checkSettingValue(settings, "login_captcha")),
			RegCaptcha:    model.IsTrueVal(checkSettingValue(settings, "reg_captcha")),
			ForgetCaptcha: model.IsTrueVal(checkSettingValue(settings, "forget_captcha")),
			EmailActive:   model.IsTrueVal(checkSettingValue(settings, "email_active")),
			QQLogin:       model.IsTrueVal(checkSettingValue(settings, "qq_login")),
			Themes:        checkSettingValue(settings, "themes"),
			DefaultTheme:  checkSettingValue(settings, "defaultTheme"),
		}}
}
