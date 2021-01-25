package middleware

import (
	"github.com/cloudreve/Cloudreve/v3/bootstrap"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strings"
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
		util.Log().Warning("静态文件[index.html]不存在，可能会影响首页展示")
		return ignoreFunc
	}

	fileContentBytes, err := ioutil.ReadAll(file)
	if err != nil {
		util.Log().Warning("静态文件[index.html]读取失败，可能会影响首页展示")
		return ignoreFunc
	}
	fileContent := string(fileContentBytes)

	fileServer := http.FileServer(bootstrap.StaticFS)
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// API 跳过
		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/custom") || strings.HasPrefix(path, "/dav") || path == "/manifest.json" {
			c.Next()
			return
		}

		// 不存在的路径和index.html均返回index.html
		if (path == "/index.html") || (path == "/") || !bootstrap.StaticFS.Exists("/", path) {
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

		// 存在的静态文件
		fileServer.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}
