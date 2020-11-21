package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSession(t *testing.T) {
	asserts := assert.New(t)

	{
		handler := Session("2333")
		asserts.NotNil(handler)
		asserts.NotNil(Store)
		asserts.IsType(emptyFunc(), handler)
	}
	{
		conf.RedisConfig.Server = "123"
		asserts.Panics(func() {
			Session("2333")
		})
		conf.RedisConfig.Server = ""
	}

}

func emptyFunc() gin.HandlerFunc {
	return func(c *gin.Context) {}
}

func TestCSRFInit(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()
	sessionFunc := Session("233")
	{
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest("GET", "/test", nil)
		sessionFunc(c)
		CSRFInit()(c)
		asserts.True(util.GetSession(c, "CSRF").(bool))
	}
}

func TestCSRFCheck(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()
	sessionFunc := Session("233")

	// 通过检查
	{
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest("GET", "/test", nil)
		sessionFunc(c)
		CSRFInit()(c)
		CSRFCheck()(c)
		asserts.False(c.IsAborted())
	}

	// 未通过检查
	{
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest("GET", "/test", nil)
		sessionFunc(c)
		CSRFCheck()(c)
		asserts.True(c.IsAborted())
	}
}
