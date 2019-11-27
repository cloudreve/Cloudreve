package middleware

import (
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
)

// Store session存储
var Store memstore.Store

// Session 初始化session
func Session(secret string) gin.HandlerFunc {
	Store, err := redis.NewStore(10, "tcp", "127.0.0.1:6379", "52121225", []byte("secret"))
	if err != nil {
		util.Log().Panic("无法连接到Redis：%s", err)
	}
	// Also set Secure: true if using SSL, you should though
	Store.Options(sessions.Options{HttpOnly: true, MaxAge: 7 * 86400, Path: "/"})
	return sessions.Sessions("cloudreve-session", Store)
}
