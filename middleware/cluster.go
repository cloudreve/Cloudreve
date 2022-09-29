package middleware

import (
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/cluster"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/gin-gonic/gin"
	"strconv"
)

// MasterMetadata 解析主机节点发来请求的包含主机节点信息的元数据
func MasterMetadata() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("MasterSiteID", c.GetHeader(auth.CrHeaderPrefix+"Site-Id"))
		c.Set("MasterSiteURL", c.GetHeader(auth.CrHeaderPrefix+"Site-Url"))
		c.Set("MasterVersion", c.GetHeader(auth.CrHeaderPrefix+"Cloudreve-Version"))
		c.Next()
	}
}

// UseSlaveAria2Instance 从机用于获取对应主机节点的Aria2实例
func UseSlaveAria2Instance(clusterController cluster.Controller) gin.HandlerFunc {
	return func(c *gin.Context) {
		if siteID, exist := c.Get("MasterSiteID"); exist {
			// 获取对应主机节点的从机Aria2实例
			caller, err := clusterController.GetAria2Instance(siteID.(string))
			if err != nil {
				c.JSON(200, serializer.Err(serializer.CodeNotSet, "Failed to get Aria2 instance", err))
				c.Abort()
				return
			}

			c.Set("MasterAria2Instance", caller)
			c.Next()
			return
		}

		c.JSON(200, serializer.ParamErr("Unknown master node ID", nil))
		c.Abort()
	}
}

func SlaveRPCSignRequired(nodePool cluster.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		nodeID, err := strconv.ParseUint(c.GetHeader(auth.CrHeaderPrefix+"Node-Id"), 10, 64)
		if err != nil {
			c.JSON(200, serializer.ParamErr("Unknown master node ID", err))
			c.Abort()
			return
		}

		slaveNode := nodePool.GetNodeByID(uint(nodeID))
		if slaveNode == nil {
			c.JSON(200, serializer.ParamErr("Unknown master node ID", err))
			c.Abort()
			return
		}

		SignRequired(slaveNode.MasterAuthInstance())(c)

	}
}
