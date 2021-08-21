package middleware

import (
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/slave"
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

// UseSlaveAria2Instance 从机用于获取对应主机节点的Aria2实例
func UseSlaveAria2Instance() gin.HandlerFunc {
	return func(c *gin.Context) {
		if siteID, exist := c.Get("MasterSiteID"); exist {
			// 获取对应主机节点的从机Aria2实例
			caller, err := slave.DefaultController.GetAria2Instance(siteID.(string))
			if err != nil {
				c.JSON(200, serializer.Err(serializer.CodeNotSet, "无法获取 Aria2 实例", err))
				c.Abort()
				return
			}

			c.Set("MasterAria2Instance", caller)
			c.Next()
			return
		}

		c.JSON(200, serializer.ParamErr("未知的主机节点ID", nil))
		c.Abort()
	}
}
