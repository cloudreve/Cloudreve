package routers

import (
	"bytes"
	"cloudreve/models"
	"cloudreve/pkg/serializer"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var mock sqlmock.Sqlmock

// TestMain 初始化数据库Mock
func TestMain(m *testing.M) {
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

func TestUserSession(t *testing.T) {
	asserts := assert.New(t)
	router := InitRouter()
	w := httptest.NewRecorder()

	testCases := []struct {
		settingRows *sqlmock.Rows
		userRows    *sqlmock.Rows
		reqBody     string
		expected    interface{}
	}{
		// 登录信息正确，不需要验证码
		{
			settingRows: sqlmock.NewRows([]string{"name", "value", "type"}).
				AddRow("login_captcha", "0", "login"),
			userRows: sqlmock.NewRows([]string{"email", "nick", "password", "options"}).
				AddRow("admin@cloudreve.org", "admin", "CKLmDKa1C9SD64vU:76adadd4fd4bad86959155f6f7bc8993c94e7adf", "{}"),
			reqBody: `{"userName":"admin@cloudreve.org","captchaCode":"captchaCode","Password":"admin"}`,
			expected: serializer.BuildUserResponse(model.User{
				Email: "admin@cloudreve.org",
				Nick:  "admin",
			}),
		},
		// 邮箱正确密码错误
		{
			settingRows: sqlmock.NewRows([]string{"name", "value", "type"}).
				AddRow("login_captcha", "0", "login"),
			userRows: sqlmock.NewRows([]string{"email", "nick", "password", "options"}).
				AddRow("admin@cloudreve.org", "admin", "CKLmDKa1C9SD64vU:76adadd4fd4bad86959155f6f7bc8993c94e7adf", "{}"),
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
			reqBody:  `{"userName":"admin@cloudreve.org","captchaCode":"captchaCode","Password":"admin"}`,
			expected: serializer.Err(403, "该账号已被封禁", nil),
		},
		// 用户未激活
		{
			settingRows: sqlmock.NewRows([]string{"name", "value", "type"}).
				AddRow("login_captcha", "0", "login"),
			userRows: sqlmock.NewRows([]string{"email", "nick", "password", "options", "status"}).
				AddRow("admin@cloudreve.org", "admin", "CKLmDKa1C9SD64vU:76adadd4fd4bad86959155f6f7bc8993c94e7adf", "{}", model.NotActivicated),
			reqBody:  `{"userName":"admin@cloudreve.org","captchaCode":"captchaCode","Password":"admin"}`,
			expected: serializer.Err(403, "该账号未激活", nil),
		},
	}

	for _, testCase := range testCases {
		if testCase.settingRows != nil {
			mock.ExpectQuery("^SELECT (.+)").WillReturnRows(testCase.settingRows)
		}
		if testCase.userRows != nil {
			mock.ExpectQuery("^SELECT (.+)").WillReturnRows(testCase.userRows)
		}
		req, _ := http.NewRequest(
			"POST",
			"/Api/V3/User/Session",
			bytes.NewReader([]byte(testCase.reqBody)),
		)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		expectedJson, _ := json.Marshal(testCase.expected)
		asserts.JSONEq(string(expectedJson), w.Body.String())

		w.Body.Reset()
		asserts.NoError(mock.ExpectationsWereMet())
		model.ClearCache()
	}

}
