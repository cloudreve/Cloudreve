package middleware

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/cloudreve/Cloudreve/v3/bootstrap"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-gonic/gin"
)

// FrontendFileHandler 前端静态文件处理
func FrontendFileHandler() gin.HandlerFunc {
	ignoreFunc := func(c *gin.Context) {
		c.Next()
	}

	if bootstrap.StaticFS == nil {
		return ignoreFunc
	}

	// 读取index.html
	file, err := bootstrap.StaticFS.Open("/index.html")
	if err != nil {
		util.Log().Warning("Static file \"index.html\" does not exist, it might affect the display of the homepage.")
		return ignoreFunc
	}

	fileContentBytes, err := ioutil.ReadAll(file)
	if err != nil {
		util.Log().Warning("Cannot read static file \"index.html\", it might affect the display of the homepage.")
		return ignoreFunc
	}
	fileContent := string(fileContentBytes)

	fileServer := http.FileServer(bootstrap.StaticFS)
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// API 跳过
		if strings.HasPrefix(path, "/api") ||
			strings.HasPrefix(path, "/custom") ||
			strings.HasPrefix(path, "/dav") ||
			strings.HasPrefix(path, "/f") ||
			path == "/manifest.json" {
			c.Next()
			return
		}

		// 不存在的路径和index.html均返回index.html
		if (path == "/index.html") || (path == "/") || !bootstrap.StaticFS.Exists("/", path) {
			// 读取、替换站点设置
			options := model.GetSettingByNames(
				"siteName",       // 站点名称
				"siteKeywords",   // 关键词
				"siteDes",        // 描述
				"siteScript",     // 自定义代码
				"pwa_small_icon", // 图标
			)
			finalHTML := util.Replace(map[string]string{
				"{siteName}":       options["siteName"],
				"{siteKeywords}":   options["siteKeywords"],
				"{siteDes}":        options["siteDes"],
				"{siteScript}":     options["siteScript"],
				"{pwa_small_icon}": options["pwa_small_icon"],
			}, fileContent)

			c.Header("Content-Type", "text/html")
			c.String(200, finalHTML)
			c.Abort()
			return
		}

		if path == "/service-worker.js" {
			c.Header("Cache-Control", "public, no-cache")
		}

		// 存在的静态文件
		fileServer.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}
