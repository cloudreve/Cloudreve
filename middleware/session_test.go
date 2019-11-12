package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSession(t *testing.T) {
	asserts := assert.New(t)

	handler := Session("2333")
	asserts.NotNil(handler)
	asserts.NotNil(Store)
	asserts.IsType(emptyFunc(), handler)
}

func emptyFunc() gin.HandlerFunc {
	return func(c *gin.Context) {}
}
