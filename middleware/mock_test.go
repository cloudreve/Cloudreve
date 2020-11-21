package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMockHelper(t *testing.T) {
	asserts := assert.New(t)
	MockHelperFunc := MockHelper()
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request, _ = http.NewRequest("GET", "/test", nil)

	// 写入session
	{
		SessionMock["test"] = "pass"
		Session("test")(c)
		MockHelperFunc(c)
		asserts.Equal("pass", util.GetSession(c, "test").(string))
	}

	// 写入context
	{
		ContextMock["test"] = "pass"
		MockHelperFunc(c)
		test, exist := c.Get("test")
		asserts.True(exist)
		asserts.Equal("pass", test.(string))

	}
}
