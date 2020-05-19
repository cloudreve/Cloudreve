package middleware

import (
	"errors"
	"github.com/HFO4/cloudreve/bootstrap"
	"github.com/HFO4/cloudreve/pkg/cache"
	"github.com/HFO4/cloudreve/pkg/hashid"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
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

type StaticMock struct {
	testMock.Mock
}

func (m StaticMock) Open(name string) (http.File, error) {
	args := m.Called(name)
	return args.Get(0).(http.File), args.Error(1)
}

func (m StaticMock) Exists(prefix string, filepath string) bool {
	args := m.Called(prefix, filepath)
	return args.Bool(0)
}

func TestInjectSiteInfo(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()

	// 静态资源未加载
	{
		TestFunc := InjectSiteInfo()

		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{}
		c.Request, _ = http.NewRequest("GET", "/", nil)
		TestFunc(c)
		asserts.False(c.IsAborted())
	}

	// index.html 不存在
	{
		testStatic := &StaticMock{}
		bootstrap.StaticFS = testStatic
		testStatic.On("Open", "/index.html").
			Return(&os.File{}, errors.New("error"))
		TestFunc := InjectSiteInfo()

		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{}
		c.Request, _ = http.NewRequest("GET", "/", nil)
		TestFunc(c)
		asserts.False(c.IsAborted())
	}

	// index.html 读取失败
	{
		file, _ := util.CreatNestedFile("tests/index.html")
		file.Close()
		testStatic := &StaticMock{}
		bootstrap.StaticFS = testStatic
		testStatic.On("Open", "/index.html").
			Return(file, nil)
		TestFunc := InjectSiteInfo()

		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{}
		c.Request, _ = http.NewRequest("GET", "/", nil)
		TestFunc(c)
		asserts.False(c.IsAborted())
	}

	// 成功且命中
	{
		file, _ := util.CreatNestedFile("tests/index.html")
		defer file.Close()
		testStatic := &StaticMock{}
		bootstrap.StaticFS = testStatic
		testStatic.On("Open", "/index.html").
			Return(file, nil)
		TestFunc := InjectSiteInfo()

		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{}
		c.Request, _ = http.NewRequest("GET", "/", nil)

		cache.Set("setting_siteName", "cloudreve", 0)
		cache.Set("setting_siteKeywords", "cloudreve", 0)
		cache.Set("setting_siteScript", "cloudreve", 0)
		cache.Set("setting_pwa_small_icon", "cloudreve", 0)

		TestFunc(c)
		asserts.True(c.IsAborted())
	}

	// 成功且未命中
	{
		file, _ := util.CreatNestedFile("tests/index.html")
		defer file.Close()
		testStatic := &StaticMock{}
		bootstrap.StaticFS = testStatic
		testStatic.On("Open", "/index.html").
			Return(file, nil)
		TestFunc := InjectSiteInfo()

		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{}
		c.Request, _ = http.NewRequest("GET", "/2", nil)

		TestFunc(c)
		asserts.False(c.IsAborted())
	}

}
