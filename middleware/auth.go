package middleware

import (
	"github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/auth"
	"github.com/HFO4/cloudreve/pkg/cache"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

// SignRequired 验证请求签名
func SignRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		switch c.Request.Method {
		case "PUT", "POST":
			err = auth.CheckRequest(auth.General, c.Request)
			// TODO 生产环境去掉下一行
			//err = nil
		default:
			err = auth.CheckURI(auth.General, c.Request.URL)
		}

		if err != nil {
			c.JSON(200, serializer.Err(serializer.CodeCheckLogin, err.Error(), err))
			c.Abort()
			return
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

// WebDAVAuth 验证WebDAV登录及权限
func WebDAVAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// OPTIONS 请求不需要鉴权，否则Windows10下无法保存文档
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		username, password, ok := c.Request.BasicAuth()
		if !ok {
			c.Writer.Header()["WWW-Authenticate"] = []string{`Basic realm="cloudreve"`}
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}

		expectedUser, err := model.GetUserByEmail(username)
		if err != nil {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}

		// 密码正确？
		ok, _ = expectedUser.CheckPassword(password)
		if !ok {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}

		// 用户组已启用WebDAV？
		if !expectedUser.Group.WebDAVEnabled {
			c.Status(http.StatusForbidden)
			c.Abort()
			return
		}

		c.Set("user", &expectedUser)
		c.Next()
	}
}

// RemoteCallbackAuth 远程回调签名验证
// TODO 测试
func RemoteCallbackAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证 Callback Key
		callbackKey := c.Param("key")
		if callbackKey == "" {
			c.JSON(200, serializer.ParamErr("Callback Key 不能为空", nil))
			c.Abort()
			return
		}
		callbackSessionRaw, exist := cache.Get("callback_" + callbackKey)
		if !exist {
			c.JSON(200, serializer.ParamErr("回调会话不存在或已过期", nil))
			c.Abort()
			return
		}
		callbackSession := callbackSessionRaw.(serializer.UploadSession)
		c.Set("callbackSession", &callbackSession)

		// 清理回调会话
		_ = cache.Deletes([]string{callbackKey}, "callback_")

		// 查找用户
		user, err := model.GetUserByID(callbackSession.UID)
		if err != nil {
			c.JSON(200, serializer.Err(serializer.CodeCheckLogin, "找不到用户", err))
			c.Abort()
			return
		}
		c.Set("user", &user)

		// 检查存储策略是否一致
		if user.GetPolicyID() != callbackSession.PolicyID {
			c.JSON(200, serializer.Err(serializer.CodePolicyNotAllowed, "存储策略已变更，请重新上传", nil))
			c.Abort()
			return
		}

		// 验证签名
		authInstance := auth.HMACAuth{SecretKey: []byte(user.Policy.SecretKey)}
		if err := auth.CheckRequest(authInstance, c.Request); err != nil {
			c.JSON(200, serializer.Err(serializer.CodeCheckLogin, err.Error(), err))
			c.Abort()
			return
		}

		c.Next()

	}
}
