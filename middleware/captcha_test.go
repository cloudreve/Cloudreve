package middleware

import (
	"bytes"
	"errors"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}

func TestCaptchaRequired_General(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()

	// 未启用验证码
	{
		cache.SetSettings(map[string]string{
			"login_captcha":                 "0",
			"captcha_type":                  "1",
			"captcha_ReCaptchaSecret":       "1",
			"captcha_TCaptcha_SecretId":     "1",
			"captcha_TCaptcha_SecretKey":    "1",
			"captcha_TCaptcha_CaptchaAppId": "1",
			"captcha_TCaptcha_AppSecretKey": "1",
		}, "setting_")
		TestFunc := CaptchaRequired("login_captcha")
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{}
		c.Request, _ = http.NewRequest("GET", "/", nil)
		TestFunc(c)
		asserts.False(c.IsAborted())
	}

	// body 无法读取
	{
		cache.SetSettings(map[string]string{
			"login_captcha":                 "1",
			"captcha_type":                  "1",
			"captcha_ReCaptchaSecret":       "1",
			"captcha_TCaptcha_SecretId":     "1",
			"captcha_TCaptcha_SecretKey":    "1",
			"captcha_TCaptcha_CaptchaAppId": "1",
			"captcha_TCaptcha_AppSecretKey": "1",
		}, "setting_")
		TestFunc := CaptchaRequired("login_captcha")
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{}
		c.Request, _ = http.NewRequest("GET", "/", errReader(1))
		TestFunc(c)
		asserts.True(c.IsAborted())
	}

	// body JSON 解析失败
	{
		cache.SetSettings(map[string]string{
			"login_captcha":                 "1",
			"captcha_type":                  "1",
			"captcha_ReCaptchaSecret":       "1",
			"captcha_TCaptcha_SecretId":     "1",
			"captcha_TCaptcha_SecretKey":    "1",
			"captcha_TCaptcha_CaptchaAppId": "1",
			"captcha_TCaptcha_AppSecretKey": "1",
		}, "setting_")
		TestFunc := CaptchaRequired("login_captcha")
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{}
		r := bytes.NewReader([]byte("123"))
		c.Request, _ = http.NewRequest("GET", "/", r)
		TestFunc(c)
		asserts.True(c.IsAborted())
	}
}

func TestCaptchaRequired_Normal(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()

	// 验证码错误
	{
		cache.SetSettings(map[string]string{
			"login_captcha":                 "1",
			"captcha_type":                  "normal",
			"captcha_ReCaptchaSecret":       "1",
			"captcha_TCaptcha_SecretId":     "1",
			"captcha_TCaptcha_SecretKey":    "1",
			"captcha_TCaptcha_CaptchaAppId": "1",
			"captcha_TCaptcha_AppSecretKey": "1",
		}, "setting_")
		TestFunc := CaptchaRequired("login_captcha")
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{}
		r := bytes.NewReader([]byte("{}"))
		c.Request, _ = http.NewRequest("GET", "/", r)
		Session("233")(c)
		TestFunc(c)
		asserts.True(c.IsAborted())
	}
}

func TestCaptchaRequired_Recaptcha(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()

	// 无法初始化reCaptcha实例
	{
		cache.SetSettings(map[string]string{
			"login_captcha":                 "1",
			"captcha_type":                  "recaptcha",
			"captcha_ReCaptchaSecret":       "",
			"captcha_TCaptcha_SecretId":     "1",
			"captcha_TCaptcha_SecretKey":    "1",
			"captcha_TCaptcha_CaptchaAppId": "1",
			"captcha_TCaptcha_AppSecretKey": "1",
		}, "setting_")
		TestFunc := CaptchaRequired("login_captcha")
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{}
		r := bytes.NewReader([]byte("{}"))
		c.Request, _ = http.NewRequest("GET", "/", r)
		TestFunc(c)
		asserts.True(c.IsAborted())
	}

	// 验证码错误
	{
		cache.SetSettings(map[string]string{
			"login_captcha":                 "1",
			"captcha_type":                  "recaptcha",
			"captcha_ReCaptchaSecret":       "233",
			"captcha_TCaptcha_SecretId":     "1",
			"captcha_TCaptcha_SecretKey":    "1",
			"captcha_TCaptcha_CaptchaAppId": "1",
			"captcha_TCaptcha_AppSecretKey": "1",
		}, "setting_")
		TestFunc := CaptchaRequired("login_captcha")
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{}
		r := bytes.NewReader([]byte("{}"))
		c.Request, _ = http.NewRequest("GET", "/", r)
		TestFunc(c)
		asserts.True(c.IsAborted())
	}
}

func TestCaptchaRequired_Tcaptcha(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()

	// 验证出错
	{
		cache.SetSettings(map[string]string{
			"login_captcha":                 "1",
			"captcha_type":                  "tcaptcha",
			"captcha_ReCaptchaSecret":       "",
			"captcha_TCaptcha_SecretId":     "1",
			"captcha_TCaptcha_SecretKey":    "1",
			"captcha_TCaptcha_CaptchaAppId": "1",
			"captcha_TCaptcha_AppSecretKey": "1",
		}, "setting_")
		TestFunc := CaptchaRequired("login_captcha")
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{}
		r := bytes.NewReader([]byte("{}"))
		c.Request, _ = http.NewRequest("GET", "/", r)
		TestFunc(c)
		asserts.True(c.IsAborted())
	}
}
