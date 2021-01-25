package middleware

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/gin-gonic/gin"
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
