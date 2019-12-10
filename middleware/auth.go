package middleware

import (
	"github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/auth"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// SignRequired 验证请求签名
// TODO 测试
func SignRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := auth.CheckURI(c.Request.URL)
		if err != nil {
			c.JSON(200, serializer.Err(serializer.CodeCheckLogin, err.Error(), err))
			c.Abort()
		}
		c.Next()
	}
}

// CurrentUser 获取登录用户
func CurrentUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		uid := session.Get("user_id")
		if uid != nil {
			user, err := model.GetUserByID(uid)
			if err == nil {
				c.Set("user", &user)
			}
		}
		c.Next()
	}
}

// AuthRequired 需要登录
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if user, _ := c.Get("user"); user != nil {
			if _, ok := user.(*model.User); ok {
				c.Next()
				return
			}
		}

		c.JSON(200, serializer.CheckLogin())
		c.Abort()
	}
}
