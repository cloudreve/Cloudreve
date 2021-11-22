package model

import (
	"context"
	"github.com/cloudreve/Cloudreve/v3/models/scripts/invoker"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/fatih/color"
	"github.com/gofrs/uuid"
	"github.com/hashicorp/go-version"
	"github.com/jinzhu/gorm"
	"sort"
	"strings"
)

// 是否需要迁移
func needMigration() bool {
	var setting Setting
	return DB.Where("name = ?", "db_version_"+conf.RequiredDBVersion).First(&setting).Error != nil
}

//执行数据迁移
func migration() {
	// 确认是否需要执行迁移
	if !needMigration() {
		util.Log().Info("数据库版本匹配，跳过数据库迁移")
		return

	}

	util.Log().Info("开始进行数据库初始化...")

	// 清除所有缓存
	if instance, ok := cache.Store.(*cache.RedisStore); ok {
		instance.DeleteAll()
	}

	// 自动迁移模式
	if conf.DatabaseConfig.Type == "mysql" {
		DB = DB.Set("gorm:table_options", "ENGINE=InnoDB")
	}

	DB.AutoMigrate(&User{}, &Setting{}, &Group{}, &Policy{}, &Folder{}, &File{}, &Share{},
		&Task{}, &Download{}, &Tag{}, &Webdav{}, &Node{})

	// 创建初始存储策略
	addDefaultPolicy()

	// 创建初始用户组
	addDefaultGroups()

	// 创建初始管理员账户
	addDefaultUser()

	// 创建初始节点
	addDefaultNode()

	// 向设置数据表添加初始设置
	addDefaultSettings()

	// 执行数据库升级脚本
	execUpgradeScripts()

	util.Log().Info("数据库初始化结束")

}

func addDefaultPolicy() {
	_, err := GetPolicyByID(uint(1))
	// 未找到初始存储策略时，则创建
	if gorm.IsRecordNotFoundError(err) {
		defaultPolicy := Policy{
			Name:               "默认存储策略",
			Type:               "local",
			MaxSize:            0,
			AutoRename:         true,
			DirNameRule:        "uploads/{uid}/{path}",
			FileNameRule:       "{uid}_{randomkey8}_{originname}",
			IsOriginLinkEnable: false,
		}
		if err := DB.Create(&defaultPolicy).Error; err != nil {
			util.Log().Panic("无法创建初始存储策略, %s", err)
		}
	}
}

func addDefaultSettings() {
	siteID, _ := uuid.NewV4()

	defaultSettings := []Setting{
		{Name: "siteURL", Value: `http://localhost`, Type: "basic"},
		{Name: "siteName", Value: `Cloudreve`, Type: "basic"},
		{Name: "siteICPId", Value: ``, Type: "basic"},
		{Name: "register_enabled", Value: `1`, Type: "register"},
		{Name: "default_group", Value: `2`, Type: "register"},
		{Name: "siteKeywords", Value: `网盘，网盘`, Type: "basic"},
		{Name: "siteDes", Value: `Cloudreve`, Type: "basic"},
		{Name: "siteTitle", Value: `平步云端`, Type: "basic"},
		{Name: "siteScript", Value: ``, Type: "basic"},
		{Name: "siteID", Value: siteID.String(), Type: "basic"},
		{Name: "fromName", Value: `Cloudreve`, Type: "mail"},
		{Name: "mail_keepalive", Value: `30`, Type: "mail"},
		{Name: "fromAdress", Value: `no-reply@acg.blue`, Type: "mail"},
		{Name: "smtpHost", Value: `smtp.mxhichina.com`, Type: "mail"},
		{Name: "smtpPort", Value: `25`, Type: "mail"},
		{Name: "replyTo", Value: `abslant@126.com`, Type: "mail"},
		{Name: "smtpUser", Value: `no-reply@acg.blue`, Type: "mail"},
		{Name: "smtpPass", Value: ``, Type: "mail"},
		{Name: "smtpEncryption", Value: `0`, Type: "mail"},
		{Name: "maxEditSize", Value: `4194304`, Type: "file_edit"},
		{Name: "archive_timeout", Value: `60`, Type: "timeout"},
		{Name: "download_timeout", Value: `60`, Type: "timeout"},
		{Name: "preview_timeout", Value: `60`, Type: "timeout"},
		{Name: "doc_preview_timeout", Value: `60`, Type: "timeout"},
		{Name: "upload_credential_timeout", Value: `1800`, Type: "timeout"},
		{Name: "upload_session_timeout", Value: `86400`, Type: "timeout"},
		{Name: "slave_api_timeout", Value: `60`, Type: "timeout"},
		{Name: "slave_node_retry", Value: `3`, Type: "slave"},
		{Name: "slave_ping_interval", Value: `60`, Type: "slave"},
		{Name: "slave_recover_interval", Value: `120`, Type: "slave"},
		{Name: "slave_transfer_timeout", Value: `172800`, Type: "timeout"},
		{Name: "onedrive_monitor_timeout", Value: `600`, Type: "timeout"},
		{Name: "share_download_session_timeout", Value: `2073600`, Type: "timeout"},
		{Name: "onedrive_callback_check", Value: `20`, Type: "timeout"},
		{Name: "folder_props_timeout", Value: `300`, Type: "timeout"},
		{Name: "onedrive_chunk_retries", Value: `1`, Type: "retry"},
		{Name: "onedrive_source_timeout", Value: `1800`, Type: "timeout"},
		{Name: "reset_after_upload_failed", Value: `0`, Type: "upload"},
		{Name: "login_captcha", Value: `0`, Type: "login"},
		{Name: "reg_captcha", Value: `0`, Type: "login"},
		{Name: "email_active", Value: `0`, Type: "register"},
		{Name: "mail_activation_template", Value: `<!DOCTYPE html PUBLIC"-//W3C//DTD XHTML 1.0 Transitional//EN""http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd"><html xmlns="http://www.w3.org/1999/xhtml"style="font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif; box-sizing: border-box; 
font-size: 14px; margin: 0;"><head><meta name="viewport"content="width=device-width"/><meta http-equiv="Content-Type"content="text/html; charset=UTF-8"/><title>激活您的账户</title><style type="text/css">img{max-width:100%}body{-webkit-font-smoothing:antialiased;-webkit-text-size-adjust:none;width:100%!important;height:100%;line-height:1.6em}body{background-color:#f6f6f6}@media only screen and(max-width:640px){body{padding:0!important}h1{font-weight:800!important;margin:20px 0 5px!important}h2{font-weight:800!important;margin:20px 0 5px!important}h3{font-weight:800!important;margin:20px 0 5px!important}h4{font-weight:800!important;margin:20px 0 5px!important}h1{font-size:22px!important}h2{font-size:18px!important}h3{font-size:16px!important}.container{padding:0!important;width:100%!important}.content{padding:0!important}.content-wrap{padding:10px!important}.invoice{width:100%!important}}</style></head><body itemscope itemtype="http://schema.org/EmailMessage"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: 
border-box; font-size: 14px; -webkit-font-smoothing: antialiased; -webkit-text-size-adjust: none; width: 100% !important; height: 100%; line-height: 1.6em; background-color: #f6f6f6; margin: 0;"bgcolor="#f6f6f6"><table class="body-wrap"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; width: 100%; background-color: #f6f6f6; margin: 0;"bgcolor="#f6f6f6"><tr style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; 
box-sizing: border-box; font-size: 14px; margin: 0;"><td style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0;"valign="top"></td><td class="container"width="600"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; display: block !important; max-width: 600px !important; clear: both !important; margin: 0 auto;"valign="top"><div class="content"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; max-width: 600px; display: block; margin: 0 auto; padding: 20px;"><table class="main"width="100%"cellpadding="0"cellspacing="0"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; border-radius: 3px; background-color: #fff; margin: 0; border: 1px 
solid #e9e9e9;"bgcolor="#fff"><tr style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 
14px; margin: 0;"><td class="alert alert-warning"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 16px; vertical-align: top; color: #fff; font-weight: 500; text-align: center; border-radius: 3px 3px 0 0; background-color: #009688; margin: 0; padding: 20px;"align="center"bgcolor="#FF9F00"valign="top">激活{siteTitle}账户</td></tr><tr style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;"><td class="content-wrap"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0; padding: 20px;"valign="top"><table width="100%"cellpadding="0"cellspacing="0"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;"><tr style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;"><td class="content-block"style="font-family: 'Helvetica 
Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0; padding: 0 0 20px;"valign="top">亲爱的<strong style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;">{userName}</strong>：</td></tr><tr style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;"><td class="content-block"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0; padding: 0 0 20px;"valign="top">感谢您注册{siteTitle},请点击下方按钮完成账户激活。</td></tr><tr style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;"><td class="content-block"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0; padding: 0 0 20px;"valign="top"><a href="{activationUrl}"class="btn-primary"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; color: #FFF; text-decoration: none; line-height: 2em; font-weight: bold; text-align: center; cursor: pointer; display: inline-block; border-radius: 5px; text-transform: capitalize; background-color: #009688; margin: 0; border-color: #009688; border-style: solid; border-width: 10px 20px;">激活账户</a></td></tr><tr style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;"><td class="content-block"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0; padding: 0 0 20px;"valign="top">感谢您选择{siteTitle}。</td></tr></table></td></tr></table><div class="footer"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; width: 100%; clear: both; color: #999; margin: 0; padding: 20px;"><table width="100%"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;"><tr style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;"><td class="aligncenter content-block"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 12px; vertical-align: top; color: #999; text-align: center; margin: 0; padding: 0 0 20px;"align="center"valign="top">此邮件由系统自动发送，请不要直接回复。</td></tr></table></div></div></td><td style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0;"valign="top"></td></tr></table></body></html>`, Type: "mail_template"},
		{Name: "forget_captcha", Value: `0`, Type: "login"},
		{Name: "mail_reset_pwd_template", Value: `<!DOCTYPE html PUBLIC"-//W3C//DTD XHTML 1.0 Transitional//EN""http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd"><html xmlns="http://www.w3.org/1999/xhtml"style="font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif; box-sizing: border-box; 
font-size: 14px; margin: 0;"><head><meta name="viewport"content="width=device-width"/><meta http-equiv="Content-Type"content="text/html; charset=UTF-8"/><title>重设密码</title><style type="text/css">img{max-width:100%}body{-webkit-font-smoothing:antialiased;-webkit-text-size-adjust:none;width:100%!important;height:100%;line-height:1.6em}body{background-color:#f6f6f6}@media only screen and(max-width:640px){body{padding:0!important}h1{font-weight:800!important;margin:20px 0 5px!important}h2{font-weight:800!important;margin:20px 0 5px!important}h3{font-weight:800!important;margin:20px 0 5px!important}h4{font-weight:800!important;margin:20px 0 5px!important}h1{font-size:22px!important}h2{font-size:18px!important}h3{font-size:16px!important}.container{padding:0!important;width:100%!important}.content{padding:0!important}.content-wrap{padding:10px!important}.invoice{width:100%!important}}</style></head><body itemscope itemtype="http://schema.org/EmailMessage"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: 
border-box; font-size: 14px; -webkit-font-smoothing: antialiased; -webkit-text-size-adjust: none; width: 100% !important; height: 100%; line-height: 1.6em; background-color: #f6f6f6; margin: 0;"bgcolor="#f6f6f6"><table class="body-wrap"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; width: 100%; background-color: #f6f6f6; margin: 0;"bgcolor="#f6f6f6"><tr style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; 
box-sizing: border-box; font-size: 14px; margin: 0;"><td style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0;"valign="top"></td><td class="container"width="600"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; display: block !important; max-width: 600px !important; clear: both !important; margin: 0 auto;"valign="top"><div class="content"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; max-width: 600px; display: block; margin: 0 auto; padding: 20px;"><table class="main"width="100%"cellpadding="0"cellspacing="0"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; border-radius: 3px; background-color: #fff; margin: 0; border: 1px 
solid #e9e9e9;"bgcolor="#fff"><tr style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 
14px; margin: 0;"><td class="alert alert-warning"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 16px; vertical-align: top; color: #fff; font-weight: 500; text-align: center; border-radius: 3px 3px 0 0; background-color: #2196F3; margin: 0; padding: 20px;"align="center"bgcolor="#FF9F00"valign="top">重设{siteTitle}密码</td></tr><tr style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;"><td class="content-wrap"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0; padding: 20px;"valign="top"><table width="100%"cellpadding="0"cellspacing="0"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;"><tr style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;"><td class="content-block"style="font-family: 'Helvetica 
Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0; padding: 0 0 20px;"valign="top">亲爱的<strong style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;">{userName}</strong>：</td></tr><tr style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;"><td class="content-block"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0; padding: 0 0 20px;"valign="top">请点击下方按钮完成密码重设。如果非你本人操作，请忽略此邮件。</td></tr><tr style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;"><td class="content-block"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0; padding: 0 0 20px;"valign="top"><a href="{resetUrl}"class="btn-primary"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; color: #FFF; text-decoration: none; line-height: 2em; font-weight: bold; text-align: center; cursor: pointer; display: inline-block; border-radius: 5px; text-transform: capitalize; background-color: #2196F3; margin: 0; border-color: #2196F3; border-style: solid; border-width: 10px 20px;">重设密码</a></td></tr><tr style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;"><td class="content-block"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0; padding: 0 0 20px;"valign="top">感谢您选择{siteTitle}。</td></tr></table></td></tr></table><div class="footer"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; width: 100%; clear: both; color: #999; margin: 0; padding: 20px;"><table width="100%"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;"><tr style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;"><td class="aligncenter content-block"style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 12px; vertical-align: top; color: #999; text-align: center; margin: 0; padding: 0 0 20px;"align="center"valign="top">此邮件由系统自动发送，请不要直接回复。</td></tr></table></div></div></td><td style="font-family: 'Helvetica Neue',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0;"valign="top"></td></tr></table></body></html>`, Type: "mail_template"},
		{Name: "db_version_" + conf.RequiredDBVersion, Value: `installed`, Type: "version"},
		{Name: "hot_share_num", Value: `10`, Type: "share"},
		{Name: "gravatar_server", Value: `https://www.gravatar.com/`, Type: "avatar"},
		{Name: "defaultTheme", Value: `#3f51b5`, Type: "basic"},
		{Name: "themes", Value: `{"#3f51b5":{"palette":{"primary":{"main":"#3f51b5"},"secondary":{"main":"#f50057"}}},"#2196f3":{"palette":{"primary":{"main":"#2196f3"},"secondary":{"main":"#FFC107"}}},"#673AB7":{"palette":{"primary":{"main":"#673AB7"},"secondary":{"main":"#2196F3"}}},"#E91E63":{"palette":{"primary":{"main":"#E91E63"},"secondary":{"main":"#42A5F5","contrastText":"#fff"}}},"#FF5722":{"palette":{"primary":{"main":"#FF5722"},"secondary":{"main":"#3F51B5"}}},"#FFC107":{"palette":{"primary":{"main":"#FFC107"},"secondary":{"main":"#26C6DA"}}},"#8BC34A":{"palette":{"primary":{"main":"#8BC34A","contrastText":"#fff"},"secondary":{"main":"#FF8A65","contrastText":"#fff"}}},"#009688":{"palette":{"primary":{"main":"#009688"},"secondary":{"main":"#4DD0E1","contrastText":"#fff"}}},"#607D8B":{"palette":{"primary":{"main":"#607D8B"},"secondary":{"main":"#F06292"}}},"#795548":{"palette":{"primary":{"main":"#795548"},"secondary":{"main":"#4CAF50","contrastText":"#fff"}}}}`, Type: "basic"},
		{Name: "max_worker_num", Value: `10`, Type: "task"},
		{Name: "max_parallel_transfer", Value: `4`, Type: "task"},
		{Name: "secret_key", Value: util.RandStringRunes(256), Type: "auth"},
		{Name: "temp_path", Value: "temp", Type: "path"},
		{Name: "avatar_path", Value: "avatar", Type: "path"},
		{Name: "avatar_size", Value: "2097152", Type: "avatar"},
		{Name: "avatar_size_l", Value: "200", Type: "avatar"},
		{Name: "avatar_size_m", Value: "130", Type: "avatar"},
		{Name: "avatar_size_s", Value: "50", Type: "avatar"},
		{Name: "home_view_method", Value: "icon", Type: "view"},
		{Name: "share_view_method", Value: "list", Type: "view"},
		{Name: "cron_garbage_collect", Value: "@hourly", Type: "cron"},
		{Name: "authn_enabled", Value: "0", Type: "authn"},
		{Name: "captcha_type", Value: "normal", Type: "captcha"},
		{Name: "captcha_height", Value: "60", Type: "captcha"},
		{Name: "captcha_width", Value: "240", Type: "captcha"},
		{Name: "captcha_mode", Value: "3", Type: "captcha"},
		{Name: "captcha_ComplexOfNoiseText", Value: "0", Type: "captcha"},
		{Name: "captcha_ComplexOfNoiseDot", Value: "0", Type: "captcha"},
		{Name: "captcha_IsShowHollowLine", Value: "0", Type: "captcha"},
		{Name: "captcha_IsShowNoiseDot", Value: "1", Type: "captcha"},
		{Name: "captcha_IsShowNoiseText", Value: "0", Type: "captcha"},
		{Name: "captcha_IsShowSlimeLine", Value: "1", Type: "captcha"},
		{Name: "captcha_IsShowSineLine", Value: "0", Type: "captcha"},
		{Name: "captcha_CaptchaLen", Value: "6", Type: "captcha"},
		{Name: "captcha_ReCaptchaKey", Value: "defaultKey", Type: "captcha"},
		{Name: "captcha_ReCaptchaSecret", Value: "defaultSecret", Type: "captcha"},
		{Name: "captcha_TCaptcha_CaptchaAppId", Value: "", Type: "captcha"},
		{Name: "captcha_TCaptcha_AppSecretKey", Value: "", Type: "captcha"},
		{Name: "captcha_TCaptcha_SecretId", Value: "", Type: "captcha"},
		{Name: "captcha_TCaptcha_SecretKey", Value: "", Type: "captcha"},
		{Name: "thumb_width", Value: "400", Type: "thumb"},
		{Name: "thumb_height", Value: "300", Type: "thumb"},
		{Name: "pwa_small_icon", Value: "/static/img/favicon.ico", Type: "pwa"},
		{Name: "pwa_medium_icon", Value: "/static/img/logo192.png", Type: "pwa"},
		{Name: "pwa_large_icon", Value: "/static/img/logo512.png", Type: "pwa"},
		{Name: "pwa_display", Value: "standalone", Type: "pwa"},
		{Name: "pwa_theme_color", Value: "#000000", Type: "pwa"},
		{Name: "pwa_background_color", Value: "#ffffff", Type: "pwa"},
		{Name: "office_preview_service", Value: "https://view.officeapps.live.com/op/view.aspx?src={$src}", Type: "preview"},
	}

	for _, value := range defaultSettings {
		DB.Where(Setting{Name: value.Name}).Create(&value)
	}
}

func addDefaultGroups() {
	_, err := GetGroupByID(1)
	// 未找到初始管理组时，则创建
	if gorm.IsRecordNotFoundError(err) {
		defaultAdminGroup := Group{
			Name:          "管理员",
			PolicyList:    []uint{1},
			MaxStorage:    1 * 1024 * 1024 * 1024,
			ShareEnabled:  true,
			WebDAVEnabled: true,
			OptionsSerialized: GroupOption{
				ArchiveDownload: true,
				ArchiveTask:     true,
				ShareDownload:   true,
				Aria2:           true,
			},
		}
		if err := DB.Create(&defaultAdminGroup).Error; err != nil {
			util.Log().Panic("无法创建管理用户组, %s", err)
		}
	}

	err = nil
	_, err = GetGroupByID(2)
	// 未找到初始注册会员时，则创建
	if gorm.IsRecordNotFoundError(err) {
		defaultAdminGroup := Group{
			Name:          "注册会员",
			PolicyList:    []uint{1},
			MaxStorage:    1 * 1024 * 1024 * 1024,
			ShareEnabled:  true,
			WebDAVEnabled: true,
			OptionsSerialized: GroupOption{
				ShareDownload: true,
			},
		}
		if err := DB.Create(&defaultAdminGroup).Error; err != nil {
			util.Log().Panic("无法创建初始注册会员用户组, %s", err)
		}
	}

	err = nil
	_, err = GetGroupByID(3)
	// 未找到初始游客用户组时，则创建
	if gorm.IsRecordNotFoundError(err) {
		defaultAdminGroup := Group{
			Name:       "游客",
			PolicyList: []uint{},
			Policies:   "[]",
			OptionsSerialized: GroupOption{
				ShareDownload: true,
			},
		}
		if err := DB.Create(&defaultAdminGroup).Error; err != nil {
			util.Log().Panic("无法创建初始游客用户组, %s", err)
		}
	}
}

func addDefaultUser() {
	_, err := GetUserByID(1)
	password := util.RandStringRunes(8)

	// 未找到初始用户时，则创建
	if gorm.IsRecordNotFoundError(err) {
		defaultUser := NewUser()
		defaultUser.Email = "admin@cloudreve.org"
		defaultUser.Nick = "admin"
		defaultUser.Status = Active
		defaultUser.GroupID = 1
		err := defaultUser.SetPassword(password)
		if err != nil {
			util.Log().Panic("无法创建密码, %s", err)
		}
		if err := DB.Create(&defaultUser).Error; err != nil {
			util.Log().Panic("无法创建初始用户, %s", err)
		}

		c := color.New(color.FgWhite).Add(color.BgBlack).Add(color.Bold)
		util.Log().Info("初始管理员账号：" + c.Sprint("admin@cloudreve.org"))
		util.Log().Info("初始管理员密码：" + c.Sprint(password))
	}
}

func addDefaultNode() {
	_, err := GetNodeByID(1)

	if gorm.IsRecordNotFoundError(err) {
		defaultAdminGroup := Node{
			Name:   "主机（本机）",
			Status: NodeActive,
			Type:   MasterNodeType,
			Aria2OptionsSerialized: Aria2Option{
				Interval: 10,
				Timeout:  10,
			},
		}
		if err := DB.Create(&defaultAdminGroup).Error; err != nil {
			util.Log().Panic("无法创建初始节点记录, %s", err)
		}
	}
}

func execUpgradeScripts() {
	s := invoker.ListPrefix("UpgradeTo")
	versions := make([]*version.Version, len(s))
	for i, raw := range s {
		v, _ := version.NewVersion(strings.TrimPrefix(raw, "UpgradeTo"))
		versions[i] = v
	}
	sort.Sort(version.Collection(versions))

	for i := 0; i < len(versions); i++ {
		invoker.RunDBScript("UpgradeTo"+versions[i].String(), context.Background())
	}
}
