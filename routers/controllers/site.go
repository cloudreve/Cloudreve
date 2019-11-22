package controllers

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/gin-gonic/gin"
)

// SiteConfig 获取站点全局配置
func SiteConfig(c *gin.Context) {
	siteConfig := model.GetSettingByNames([]string{
		"siteName",
		"login_captcha",
		"qq_login",
		"reg_captcha",
		"email_active",
		"forget_captcha",
		"email_active",
		"themes",
		"defaultTheme",
	})

	c.JSON(200, serializer.BuildSiteConfig(siteConfig))
}
