package middleware

import (
	"github.com/gin-gonic/gin"
)

// MasterMetadata 解析主机节点发来请求的包含主机节点信息的元数据
func MasterMetadata() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("MasterSiteID", c.GetHeader("X-Site-ID"))
		c.Set("MasterSiteURL", c.GetHeader("X-Site-Ur"))
		c.Set("MasterVersion", c.GetHeader("X-Cloudreve-Version"))
		c.Next()
	}
}
