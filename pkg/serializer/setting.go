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
	Authn              bool   `json:"authn"'`
	User               User   `json:"user"`
}

type task struct {
	Status     int    `json:"status"`
	Type       int    `json:"type"`
	CreateDate string `json:"create_date"`
	Progress   int    `json:"progress"`
	Error      string `json:"error"`
}

// BuildTaskList 构建任务列表响应
func BuildTaskList(tasks []model.Task, total int) Response {
	res := make([]task, 0, len(tasks))
	for _, t := range tasks {
		res = append(res, task{
			Status:     t.Status,
			Type:       t.Type,
			CreateDate: t.CreatedAt.Format("2006-01-02 15:04:05"),
			Progress:   t.Progress,
			Error:      t.Error,
		})
	}

	return Response{Data: map[string]interface{}{
		"total": total,
		"tasks": res,
	}}
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
			Authn:              model.IsTrueVal(checkSettingValue(settings, "authn_enabled")),
			User:               userRes,
		}}
	return res
}
