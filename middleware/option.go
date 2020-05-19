package middleware

import (
	"github.com/HFO4/cloudreve/bootstrap"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/hashid"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/gin-gonic/gin"
	"io/ioutil"
)

// HashID 将给定对象的HashID转换为真实ID
func HashID(IDType int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Param("id") != "" {
			id, err := hashid.DecodeHashID(c.Param("id"), IDType)
			if err == nil {
				c.Set("object_id", id)
				c.Next()
				return
			}
			c.JSON(200, serializer.ParamErr("无法解析对象ID", nil))
			c.Abort()
			return

		}
		c.Next()
	}
}

// IsFunctionEnabled 当功能未开启时阻止访问
func IsFunctionEnabled(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !model.IsTrueVal(model.GetSettingByName(key)) {
			c.JSON(200, serializer.Err(serializer.CodeNoPermissionErr, "未开启此功能", nil))
			c.Abort()
			return
		}

		c.Next()
	}
}

// InjectSiteInfo 向首页html中插入站点信息
func InjectSiteInfo() gin.HandlerFunc {
	ignoreFunc := func(c *gin.Context) {
		c.Next()
	}
	if bootstrap.StaticFS == nil {
		return ignoreFunc
	}

	// 读取index.html
	file, err := bootstrap.StaticFS.Open("/index.html")
	if err != nil {
		util.Log().Warning("静态文件[index.html]不存在，可能会影响首页展示")
		return ignoreFunc
	}

	fileContentBytes, err := ioutil.ReadAll(file)
	if err != nil {
		util.Log().Warning("静态文件[index.html]读取失败，可能会影响首页展示")
		return ignoreFunc
	}
	fileContent := string(fileContentBytes)

	return func(c *gin.Context) {
		if c.Request.URL.Path == "/" || c.Request.URL.Path == "/index.html" {
			// 读取、替换站点设置
			options := model.GetSettingByNames("siteName", "siteKeywords", "siteScript",
				"pwa_small_icon")
			finalHTML := util.Replace(map[string]string{
				"{siteName}":       options["siteName"],
				"{siteDes}":        options["siteDes"],
				"{siteScript}":     options["siteScript"],
				"{pwa_small_icon}": options["pwa_small_icon"],
			}, fileContent)

			c.Header("Content-Type", "text/html")
			c.String(200, finalHTML)
			c.Abort()
			return
		}
		c.Next()
	}
}
