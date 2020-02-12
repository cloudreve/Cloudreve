package serializer

import model "github.com/HFO4/cloudreve/models"

// SiteConfig 站点全局设置序列
type SiteConfig struct {
	SiteName           string `json:"title"`
	LoginCaptcha       bool   `json:"loginCaptcha"`
	RegCaptcha         bool   `json:"regCaptcha"`
	ForgetCaptcha      bool   `json:"forgetCaptcha"`
	EmailActive        bool   `json:"emailActive"`
	QQLogin            bool   `json:"QQLogin"`
	Themes             string `json:"themes"`
	DefaultTheme       string `json:"defaultTheme"`
	ScoreEnabled       bool   `json:"score_enabled"`
	ShareScoreRate     string `json:"share_score_rate"`
	HomepageViewMethod string `json:"home_view_method"`
	ShareViewMethod    string `json:"share_view_method"`
	User               User   `json:"user"`
}

func checkSettingValue(setting map[string]string, key string) string {
	if v, ok := setting[key]; ok {
		return v
	}
	return ""
}

// BuildSiteConfig 站点全局设置
func BuildSiteConfig(settings map[string]string, user *model.User) Response {
	var userRes User
	if user != nil {
		userRes = BuildUser(*user)
	} else {
		userRes = BuildUser(*model.NewAnonymousUser())
	}
	res := Response{
		Data: SiteConfig{
			SiteName:           checkSettingValue(settings, "siteName"),
			LoginCaptcha:       model.IsTrueVal(checkSettingValue(settings, "login_captcha")),
			RegCaptcha:         model.IsTrueVal(checkSettingValue(settings, "reg_captcha")),
			ForgetCaptcha:      model.IsTrueVal(checkSettingValue(settings, "forget_captcha")),
			EmailActive:        model.IsTrueVal(checkSettingValue(settings, "email_active")),
			QQLogin:            model.IsTrueVal(checkSettingValue(settings, "qq_login")),
			Themes:             checkSettingValue(settings, "themes"),
			DefaultTheme:       checkSettingValue(settings, "defaultTheme"),
			ScoreEnabled:       model.IsTrueVal(checkSettingValue(settings, "score_enabled")),
			ShareScoreRate:     checkSettingValue(settings, "share_score_rate"),
			HomepageViewMethod: checkSettingValue(settings, "home_view_method"),
			ShareViewMethod:    checkSettingValue(settings, "share_view_method"),
			User:               userRes,
		}}
	return res
}
