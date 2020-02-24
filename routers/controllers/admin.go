package controllers

import (
	"github.com/HFO4/cloudreve/pkg/email"
	"github.com/HFO4/cloudreve/pkg/request"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/service/admin"
	"github.com/gin-gonic/gin"
	"io"
)

// AdminSummary 获取管理站点概况
func AdminSummary(c *gin.Context) {
	var service admin.NoParamService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.Summary()
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// AdminNews 获取社区新闻
func AdminNews(c *gin.Context) {
	r := request.HTTPClient{}
	res := r.Request("GET", "https://forum.cloudreve.org/api/discussions?include=startUser%2ClastUser%2CstartPost%2Ctags&filter%5Bq%5D=%20tag%3Anotice&sort=-startTime&", nil)
	io.Copy(c.Writer, res.Response.Body)
}

// AdminChangeSetting 获取站点设定项
func AdminChangeSetting(c *gin.Context) {
	var service admin.BatchSettingChangeService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Change()
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// AdminGetSetting 获取站点设置
func AdminGetSetting(c *gin.Context) {
	var service admin.BatchSettingGet
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Get()
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// AdminGetGroups 获取用户组列表
func AdminGetGroups(c *gin.Context) {
	var service admin.NoParamService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.GroupList()
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// AdminReloadService 重新加载子服务
func AdminReloadService(c *gin.Context) {
	service := c.Param("service")
	switch service {
	case "email":
		email.Init()
	}

	c.JSON(200, serializer.Response{})
}

// AdminSendTestMail 发送测试邮件
func AdminSendTestMail(c *gin.Context) {
	var service admin.MailTestService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Send()
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// AdminListRedeems 列出激活码
func AdminListRedeems(c *gin.Context) {
	var service admin.AdminListService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Redeems()
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// AdminGenerateRedeems 生成激活码
func AdminGenerateRedeems(c *gin.Context) {
	var service admin.GenerateRedeemsService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Generate()
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// AdminDeleteRedeem 删除激活码
func AdminDeleteRedeem(c *gin.Context) {
	var service admin.SingleIDService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.DeleteRedeem()
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}
