package email

import (
	"fmt"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

// NewOveruseNotification 新建超额提醒邮件
func NewOveruseNotification(userName, reason string) (string, string) {
	options := model.GetSettingByNames("siteName", "siteURL", "siteTitle", "over_used_template")
	replace := map[string]string{
		"{siteTitle}":    options["siteName"],
		"{userName}":     userName,
		"{notifyReason}": reason,
		"{siteUrl}":      options["siteURL"],
		"{siteSecTitle}": options["siteTitle"],
	}
	return fmt.Sprintf("【%s】空间容量超额提醒", options["siteName"]),
		util.Replace(replace, options["over_used_template"])
}

// NewActivationEmail 新建激活邮件
func NewActivationEmail(userName, activateURL string) (string, string) {
	options := model.GetSettingByNames("siteName", "siteURL", "siteTitle", "mail_activation_template")
	replace := map[string]string{
		"{siteTitle}":     options["siteName"],
		"{userName}":      userName,
		"{activationUrl}": activateURL,
		"{siteUrl}":       options["siteURL"],
		"{siteSecTitle}":  options["siteTitle"],
	}
	return fmt.Sprintf("【%s】注册激活", options["siteName"]),
		util.Replace(replace, options["mail_activation_template"])
}

// NewResetEmail 新建重设密码邮件
func NewResetEmail(userName, resetURL string) (string, string) {
	options := model.GetSettingByNames("siteName", "siteURL", "siteTitle", "mail_reset_pwd_template")
	replace := map[string]string{
		"{siteTitle}":    options["siteName"],
		"{userName}":     userName,
		"{resetUrl}":     resetURL,
		"{siteUrl}":      options["siteURL"],
		"{siteSecTitle}": options["siteTitle"],
	}
	return fmt.Sprintf("【%s】密码重置", options["siteName"]),
		util.Replace(replace, options["mail_reset_pwd_template"])
}
