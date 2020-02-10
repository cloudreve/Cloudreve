package middleware

import (
	"github.com/HFO4/cloudreve/pkg/hashid"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/gin-gonic/gin"
)

// HashID 将给定文件的HashID转换为真实ID
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
