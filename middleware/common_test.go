package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/hashid"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHashID(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()
	TestFunc := HashID(hashid.FolderID)

	// 未给定ID对象，跳过
	{
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{}
		c.Request, _ = http.NewRequest("POST", "/api/v3/file/dellete/1", nil)
		TestFunc(c)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.False(c.IsAborted())
	}

	// 给定ID，解析失败
	{
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"id", "2333"},
		}
		c.Request, _ = http.NewRequest("POST", "/api/v3/file/dellete/1", nil)
		TestFunc(c)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.True(c.IsAborted())
	}

	// 给定ID，解析成功
	{
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"id", hashid.HashID(1, hashid.FolderID)},
		}
		c.Request, _ = http.NewRequest("POST", "/api/v3/file/dellete/1", nil)
		TestFunc(c)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.False(c.IsAborted())
	}
}

func TestIsFunctionEnabled(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()
	TestFunc := IsFunctionEnabled("TestIsFunctionEnabled")

	// 未开启
	{
		cache.Set("setting_TestIsFunctionEnabled", "0", 0)
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{}
		c.Request, _ = http.NewRequest("POST", "/api/v3/file/dellete/1", nil)
		TestFunc(c)
		asserts.True(c.IsAborted())
	}
	// 开启
	{
		cache.Set("setting_TestIsFunctionEnabled", "1", 0)
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{}
		c.Request, _ = http.NewRequest("POST", "/api/v3/file/dellete/1", nil)
		TestFunc(c)
		asserts.False(c.IsAborted())
	}

}
