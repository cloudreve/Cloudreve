package middleware

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/wopi"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

const (
	WopiSessionCtx = "wopi_session"
)

// WopiWriteAccess validates if write access is obtained.
func WopiWriteAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := c.MustGet(WopiSessionCtx).(*wopi.SessionCache)
		if session.Action != wopi.ActionEdit {
			c.Status(http.StatusNotFound)
			c.Header(wopi.ServerErrorHeader, "read-only access")
			c.Abort()
			return
		}

		c.Next()
	}
}

func WopiAccessValidation(w wopi.Client, store cache.Driver) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken := strings.Split(c.Query(wopi.AccessTokenQuery), ".")
		if len(accessToken) != 2 {
			c.Status(http.StatusForbidden)
			c.Header(wopi.ServerErrorHeader, "malformed access token")
			c.Abort()
			return
		}

		sessionRaw, exist := store.Get(wopi.SessionCachePrefix + accessToken[0])
		if !exist {
			c.Status(http.StatusForbidden)
			c.Header(wopi.ServerErrorHeader, "invalid access token")
			c.Abort()
			return
		}

		session := sessionRaw.(wopi.SessionCache)
		user, err := model.GetActiveUserByID(session.UserID)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			c.Header(wopi.ServerErrorHeader, "user not found")
			c.Abort()
			return
		}

		fileID := c.MustGet("object_id").(uint)
		if fileID != session.FileID {
			c.Status(http.StatusInternalServerError)
			c.Header(wopi.ServerErrorHeader, "file not found")
			c.Abort()
			return
		}

		c.Set("user", &user)
		c.Set(WopiSessionCtx, &session)
		c.Next()
	}
}
