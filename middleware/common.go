package middleware

import (
	"fmt"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/gin-gonic/gin"
	"net/http"
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
			c.JSON(200, serializer.ParamErr("Failed to parse object ID", nil))
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
			c.JSON(200, serializer.Err(serializer.CodeFeatureNotEnabled, "This feature is not enabled", nil))
			c.Abort()
			return
		}

		c.Next()
	}
}

// CacheControl 屏蔽客户端缓存
func CacheControl() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "private, no-cache")
	}
}

func Sandbox() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Security-Policy", "sandbox")
	}
}

// StaticResourceCache 使用静态资源缓存策略
func StaticResourceCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", model.GetIntSetting("public_resource_maxage", 86400)))

	}
}

// MobileRequestOnly
func MobileRequestOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader(auth.CrHeaderPrefix+"ios") == "" {
			c.Redirect(http.StatusMovedPermanently, model.GetSiteURL().String())
			c.Abort()
			return
		}

		c.Next()
	}
}
