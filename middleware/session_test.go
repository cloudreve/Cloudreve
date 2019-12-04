package middleware

import (
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"testing"
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
	}

}

func emptyFunc() gin.HandlerFunc {
	return func(c *gin.Context) {}
}
