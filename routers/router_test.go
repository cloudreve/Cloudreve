package routers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/HFO4/cloudreve/middleware"
	"github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/mojocn/base64Captcha"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var mock sqlmock.Sqlmock

// TestMain 初始化数据库Mock
func TestMain(m *testing.M) {
	// 设置gin为测试模式
	gin.SetMode(gin.TestMode)

	var db *sql.DB
	var err error
	db, mock, err = sqlmock.New()
	if err != nil {
		panic("An error was not expected when opening a stub database connection")
	}
	model.DB, _ = gorm.Open("mysql", db)
	defer db.Close()

	m.Run()
}

func TestPing(t *testing.T) {
	asserts := assert.New(t)
	router := InitRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/Api/V3/Ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	asserts.Contains(w.Body.String(), "Pong")
}

func TestCaptcha(t *testing.T) {
	asserts := assert.New(t)
	router := InitRouter()
	w := httptest.NewRecorder()

	req, _ := http.NewRequest(
		"GET",
		"/Api/V3/Captcha",
		nil,
	)

	router.ServeHTTP(w, req)

	asserts.Equal(200, w.Code)
	asserts.Contains(w.Body.String(), "base64")
}

func TestUserSession(t *testing.T) {
	asserts := assert.New(t)
	router := InitRouter()
	w := httptest.NewRecorder()

	// 创建测试用验证码
	var configD = base64Captcha.ConfigDigit{
		Height:     80,
		Width:      240,
		MaxSkew:    0.7,
		DotCount:   80,
		CaptchaLen: 1,
	}
	idKeyD, _ := base64Captcha.GenerateCaptcha("", configD)
	middleware.ContextMock = map[string]interface{}{
		"captchaID": idKeyD,
	}

	testCases := []struct {
		settingRows *sqlmock.Rows
		userRows    *sqlmock.Rows
		policyRows  *sqlmock.Rows
		reqBody     string
		expected    interface{}
	}{
		// 登录信息正确，不需要验证码
		{
			settingRows: sqlmock.NewRows([]string{"name", "value", "type"}).
				AddRow("login_captcha", "0", "login"),
			userRows: sqlmock.NewRows([]string{"email", "nick", "password", "options"}).
				AddRow("admin@cloudreve.org", "admin", "CKLmDKa1C9SD64vU:76adadd4fd4bad86959155f6f7bc8993c94e7adf", "{}"), policyRows: sqlmock.NewRows([]string{"name", "type", "options"}).
				AddRow("默认上传策略", "local", "{\"op_name\":\"123\"}"),
			reqBody: `{"userName":"admin@cloudreve.org","captchaCode":"captchaCode","Password":"admin"}`,
			expected: serializer.BuildUserResponse(model.User{
				Email: "admin@cloudreve.org",
				Nick:  "admin",
			}),
		},
		// 登录信息正确，需要验证码,验证码错误
		{
			settingRows: sqlmock.NewRows([]string{"name", "value", "type"}).
				AddRow("login_captcha", "1", "login"),
			userRows: sqlmock.NewRows([]string{"email", "nick", "password", "options"}).
				AddRow("admin@cloudreve.org", "admin", "CKLmDKa1C9SD64vU:76adadd4fd4bad86959155f6f7bc8993c94e7adf", "{}"),
			policyRows: sqlmock.NewRows([]string{"name", "type", "options"}).
				AddRow("默认上传策略", "local", "{\"op_name\":\"123\"}"),
			reqBody:  `{"userName":"admin@cloudreve.org","captchaCode":"captchaCode","Password":"admin"}`,
			expected: serializer.ParamErr("验证码错误", nil),
		},
		// 邮箱正确密码错误
		{
			settingRows: sqlmock.NewRows([]string{"name", "value", "type"}).
				AddRow("login_captcha", "0", "login"),
			userRows: sqlmock.NewRows([]string{"email", "nick", "password", "options"}).
				AddRow("admin@cloudreve.org", "admin", "CKLmDKa1C9SD64vU:76adadd4fd4bad86959155f6f7bc8993c94e7adf", "{}"),
			policyRows: sqlmock.NewRows([]string{"name", "type", "options"}).
				AddRow("默认上传策略", "local", "{\"op_name\":\"123\"}"),
			reqBody:  `{"userName":"admin@cloudreve.org","captchaCode":"captchaCode","Password":"admin123"}`,
			expected: serializer.Err(401, "用户邮箱或密码错误", nil),
		},
		//邮箱格式不正确
		{
			reqBody:  `{"userName":"admin@cloudreve","captchaCode":"captchaCode","Password":"admin123"}`,
			expected: serializer.Err(40001, "邮箱格式不正确", errors.New("Key: 'UserLoginService.UserName' Error:Field validation for 'UserName' failed on the 'email' tag")),
		},
		// 用户被Ban
		{
			settingRows: sqlmock.NewRows([]string{"name", "value", "type"}).
				AddRow("login_captcha", "0", "login"),
			userRows: sqlmock.NewRows([]string{"email", "nick", "password", "options", "status"}).
				AddRow("admin@cloudreve.org", "admin", "CKLmDKa1C9SD64vU:76adadd4fd4bad86959155f6f7bc8993c94e7adf", "{}", model.Baned),
			policyRows: sqlmock.NewRows([]string{"name", "type", "options"}).
				AddRow("默认上传策略", "local", "{\"op_name\":\"123\"}"),
			reqBody:  `{"userName":"admin@cloudreve.org","captchaCode":"captchaCode","Password":"admin"}`,
			expected: serializer.Err(403, "该账号已被封禁", nil),
		},
		// 用户未激活
		{
			settingRows: sqlmock.NewRows([]string{"name", "value", "type"}).
				AddRow("login_captcha", "0", "login"),
			userRows: sqlmock.NewRows([]string{"email", "nick", "password", "options", "status"}).
				AddRow("admin@cloudreve.org", "admin", "CKLmDKa1C9SD64vU:76adadd4fd4bad86959155f6f7bc8993c94e7adf", "{}", model.NotActivicated),
			policyRows: sqlmock.NewRows([]string{"name", "type", "options"}).
				AddRow("默认上传策略", "local", "{\"op_name\":\"123\"}"),
			reqBody:  `{"userName":"admin@cloudreve.org","captchaCode":"captchaCode","Password":"admin"}`,
			expected: serializer.Err(403, "该账号未激活", nil),
		},
	}

	for k, testCase := range testCases {
		if testCase.settingRows != nil {
			mock.ExpectQuery("^SELECT (.+)").WillReturnRows(testCase.settingRows)
		}
		if testCase.userRows != nil {
			mock.ExpectQuery("^SELECT (.+)").WillReturnRows(testCase.userRows)
		}
		if testCase.policyRows != nil {
			mock.ExpectQuery("^SELECT \\* FROM `(.+)` WHERE `(.+)`\\.`deleted_at` IS NULL AND \\(\\(`policies`.`id` = 1\\)\\)(.+)$").WillReturnRows(testCase.policyRows)
		}
		req, _ := http.NewRequest(
			"POST",
			"/Api/V3/User/Session",
			bytes.NewReader([]byte(testCase.reqBody)),
		)
		router.ServeHTTP(w, req)

		asserts.Equal(200, w.Code)
		expectedJSON, _ := json.Marshal(testCase.expected)
		asserts.JSONEq(string(expectedJSON), w.Body.String(), "测试用例：%d", k)

		w.Body.Reset()
		asserts.NoError(mock.ExpectationsWereMet())
		model.ClearCache()
	}

}

func TestSessionAuthCheck(t *testing.T) {
	asserts := assert.New(t)
	router := InitRouter()
	w := httptest.NewRecorder()

	mock.ExpectQuery("^SELECT (.+)").WillReturnRows(sqlmock.NewRows([]string{"email", "nick", "password", "options"}).
		AddRow("admin@cloudreve.org", "admin", "CKLmDKa1C9SD64vU:76adadd4fd4bad86959155f6f7bc8993c94e7adf", "{}"))
	expectedUser, _ := model.GetUserByID(1)

	testCases := []struct {
		userRows    *sqlmock.Rows
		sessionMock map[string]interface{}
		contextMock map[string]interface{}
		expected    interface{}
	}{
		// 未登录
		{
			expected: serializer.CheckLogin(),
		},
		// 登录正常
		{
			userRows: sqlmock.NewRows([]string{"email", "nick", "password", "options"}).
				AddRow("admin@cloudreve.org", "admin", "CKLmDKa1C9SD64vU:76adadd4fd4bad86959155f6f7bc8993c94e7adf", "{}"),
			sessionMock: map[string]interface{}{"user_id": 1},
			expected:    serializer.BuildUserResponse(expectedUser),
		},
		// UID不存在
		{
			userRows:    sqlmock.NewRows([]string{"email", "nick", "password", "options"}),
			sessionMock: map[string]interface{}{"user_id": -1},
			expected:    serializer.CheckLogin(),
		},
	}

	for _, testCase := range testCases {
		req, _ := http.NewRequest(
			"GET",
			"/Api/V3/User/Me",
			nil,
		)
		if testCase.userRows != nil {
			mock.ExpectQuery("^SELECT (.+)").WillReturnRows(testCase.userRows)
		}
		middleware.ContextMock = testCase.contextMock
		middleware.SessionMock = testCase.sessionMock
		router.ServeHTTP(w, req)
		expectedJSON, _ := json.Marshal(testCase.expected)

		asserts.Equal(200, w.Code)
		asserts.JSONEq(string(expectedJSON), w.Body.String())
		asserts.NoError(mock.ExpectationsWereMet())

		w.Body.Reset()
	}

}
