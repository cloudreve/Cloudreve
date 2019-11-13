package util

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// SetSession 设置session
func SetSession(c *gin.Context, list map[string]interface{}) {
	s := sessions.Default(c)
	for key, value := range list {
		s.Set(key, value)
	}
	s.Save()
}

// GetSession 获取session
func GetSession(c *gin.Context, key string) interface{} {
	s := sessions.Default(c)
	return s.Get(key)
}

// DeleteSession 删除session
func DeleteSession(c *gin.Context, key string) {
	s := sessions.Default(c)
	s.Delete(key)
	s.Save()
}

// ClearSession 清空session
func ClearSession(c *gin.Context) {
	s := sessions.Default(c)
	s.Clear()
	s.Save()
}
