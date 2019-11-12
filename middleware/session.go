package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
)

// Store session存储
var Store memstore.Store

// Session 初始化session
func Session(secret string) gin.HandlerFunc {
	Store = memstore.NewStore([]byte(secret))
	// Also set Secure: true if using SSL, you should though
	Store.Options(sessions.Options{HttpOnly: true, MaxAge: 7 * 86400, Path: "/"})
	return sessions.Sessions("cloudreve-session", Store)
}
