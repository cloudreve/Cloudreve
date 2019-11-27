package middleware

import (
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
)

// Store session存储
var Store redis.Store

// Session 初始化session
func Session(secret string) gin.HandlerFunc {
	if conf.RedisConfig.Server != "" {
		// 如果配置使用了Redis
		var err error
		Store, err = redis.NewStoreWithDB(10, "tcp", conf.RedisConfig.Server, conf.RedisConfig.Password, conf.RedisConfig.DB, []byte(secret))
		if err != nil {
			util.Log().Panic("无法连接到Redis：%s", err)
		}
	} else {
		Store = memstore.NewStore([]byte(secret))
	}
	// Also set Secure: true if using SSL, you should though
	Store.Options(sessions.Options{HttpOnly: true, MaxAge: 7 * 86400, Path: "/"})
	return sessions.Sessions("cloudreve-session", Store)
}
