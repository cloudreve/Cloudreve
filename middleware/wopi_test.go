package middleware

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/mocks/wopimock"
	"github.com/cloudreve/Cloudreve/v3/pkg/wopi"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func TestWopiWriteAccess(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()
	testFunc := WopiWriteAccess()

	// deny preview only session
	{
		c, _ := gin.CreateTestContext(rec)
		c.Set(WopiSessionCtx, &wopi.SessionCache{Action: wopi.ActionPreview})
		testFunc(c)
		asserts.True(c.IsAborted())
	}

	// pass
	{
		c, _ := gin.CreateTestContext(rec)
		c.Set(WopiSessionCtx, &wopi.SessionCache{Action: wopi.ActionEdit})
		testFunc(c)
		asserts.False(c.IsAborted())
	}
}

func TestWopiAccessValidation(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()
	mockWopi := &wopimock.WopiClientMock{}
	mockCache := cache.NewMemoStore()
	testFunc := WopiAccessValidation(mockWopi, mockCache)

	// malformed access token
	{
		c, _ := gin.CreateTestContext(rec)
		c.AddParam(wopi.AccessTokenQuery, "000")
		testFunc(c)
		asserts.True(c.IsAborted())
	}

	// session key not exist
	{
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/wopi/files/1?access_token=", nil)
		query := c.Request.URL.Query()
		query.Set(wopi.AccessTokenQuery, "sessionID.key")
		c.Request.URL.RawQuery = query.Encode()
		testFunc(c)
		asserts.True(c.IsAborted())
	}

	// user key not exist
	{
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/wopi/files/1?access_token=", nil)
		query := c.Request.URL.Query()
		query.Set(wopi.AccessTokenQuery, "sessionID.key")
		c.Request.URL.RawQuery = query.Encode()
		mockCache.Set(wopi.SessionCachePrefix+"sessionID", wopi.SessionCache{UserID: 1, FileID: 1}, 0)
		mock.ExpectQuery("SELECT(.+)users(.+)").WillReturnError(errors.New("error"))
		testFunc(c)
		asserts.True(c.IsAborted())
		asserts.NoError(mock.ExpectationsWereMet())
	}

	// file not found
	{
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/wopi/files/1?access_token=", nil)
		query := c.Request.URL.Query()
		query.Set(wopi.AccessTokenQuery, "sessionID.key")
		c.Request.URL.RawQuery = query.Encode()
		mockCache.Set(wopi.SessionCachePrefix+"sessionID", wopi.SessionCache{UserID: 1, FileID: 1}, 0)
		mock.ExpectQuery("SELECT(.+)users(.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		c.Set("object_id", uint(0))
		testFunc(c)
		asserts.True(c.IsAborted())
		asserts.NoError(mock.ExpectationsWereMet())
	}

	// all pass
	{
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/wopi/files/1?access_token=", nil)
		query := c.Request.URL.Query()
		query.Set(wopi.AccessTokenQuery, "sessionID.key")
		c.Request.URL.RawQuery = query.Encode()
		mockCache.Set(wopi.SessionCachePrefix+"sessionID", wopi.SessionCache{UserID: 1, FileID: 1}, 0)
		mock.ExpectQuery("SELECT(.+)users(.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		c.Set("object_id", uint(1))
		testFunc(c)
		asserts.False(c.IsAborted())
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NotPanics(func() {
			c.MustGet(WopiSessionCtx)
		})
		asserts.NotPanics(func() {
			c.MustGet("user")
		})
	}
}
