package middleware

import (
	"net/http"
	"strings"

	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
)

// Store session存储
var Store memstore.Store

// Session 初始化session
func Session(secret string) gin.HandlerFunc {
	// Redis设置不为空，且非测试模式时使用Redis
	if conf.RedisConfig.Server != "" && gin.Mode() != gin.TestMode {
		var err error
		Store, err = redis.NewStoreWithDB(10, conf.RedisConfig.Network, conf.RedisConfig.Server, conf.RedisConfig.Password, conf.RedisConfig.DB, []byte(secret))
		if err != nil {
			util.Log().Panic("Failed to connect to Redis：%s", err)
		}

		util.Log().Info("Connect to Redis server %q.", conf.RedisConfig.Server)
	} else {
		Store = memstore.NewStore([]byte(secret))
	}

	sameSiteMode := http.SameSiteDefaultMode
	switch strings.ToLower(conf.CORSConfig.SameSite) {
	case "default":
		sameSiteMode = http.SameSiteDefaultMode
	case "none":
		sameSiteMode = http.SameSiteNoneMode
	case "strict":
		sameSiteMode = http.SameSiteStrictMode
	case "lax":
		sameSiteMode = http.SameSiteLaxMode
	}

	// Also set Secure: true if using SSL, you should though
	Store.Options(sessions.Options{
		HttpOnly: true,
		MaxAge:   60 * 86400,
		Path:     "/",
		SameSite: sameSiteMode,
		Secure:   conf.CORSConfig.Secure,
	})

	return sessions.Sessions("cloudreve-session", Store)
}

// CSRFInit 初始化CSRF标记
func CSRFInit() gin.HandlerFunc {
	return func(c *gin.Context) {
		util.SetSession(c, map[string]interface{}{"CSRF": true})
		c.Next()
	}
}

// CSRFCheck 检查CSRF标记
func CSRFCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		if check, ok := util.GetSession(c, "CSRF").(bool); ok && check {
			c.Next()
			return
		}

		c.JSON(200, serializer.Err(serializer.CodeNoPermissionErr, "Invalid origin", nil))
		c.Abort()
	}
}
