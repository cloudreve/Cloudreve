package middleware

import (
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-gonic/gin"
)

// SessionMock 测试时模拟Session
var SessionMock = make(map[string]interface{})

// ContextMock 测试时模拟Context
var ContextMock = make(map[string]interface{})

// MockHelper 单元测试助手中间件
func MockHelper() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 将SessionMock写入会话
		util.SetSession(c, SessionMock)
		for key, value := range ContextMock {
			c.Set(key, value)
		}
		c.Next()
	}
}
