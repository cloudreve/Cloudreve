package middleware

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/auth"
	"github.com/HFO4/cloudreve/pkg/cache"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/qiniu/api.v7/v7/auth/qbox"
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

func TestCurrentUser(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request, _ = http.NewRequest("GET", "/test", nil)

	//session为空
	sessionFunc := Session("233")
	sessionFunc(c)
	CurrentUser()(c)
	user, _ := c.Get("user")
	asserts.Nil(user)

	//session正确
	c, _ = gin.CreateTestContext(rec)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	sessionFunc(c)
	util.SetSession(c, map[string]interface{}{"user_id": 1})
	rows := sqlmock.NewRows([]string{"id", "deleted_at", "email", "options"}).
		AddRow(1, nil, "admin@cloudreve.org", "{}")
	mock.ExpectQuery("^SELECT (.+)").WillReturnRows(rows)
	CurrentUser()(c)
	user, _ = c.Get("user")
	asserts.NotNil(user)
	asserts.NoError(mock.ExpectationsWereMet())
}

func TestAuthRequired(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	AuthRequiredFunc := AuthRequired()

	// 未登录
	AuthRequiredFunc(c)
	asserts.NotNil(c)

	// 类型错误
	c.Set("user", 123)
	AuthRequiredFunc(c)
	asserts.NotNil(c)

	// 正常
	c.Set("user", &model.User{})
	AuthRequiredFunc(c)
	asserts.NotNil(c)
}

func TestSignRequired(t *testing.T) {
	asserts := assert.New(t)
	auth.General = auth.HMACAuth{SecretKey: []byte(util.RandStringRunes(256))}
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	SignRequiredFunc := SignRequired()

	// 鉴权失败
	SignRequiredFunc(c)
	asserts.NotNil(c)

	c.Request, _ = http.NewRequest("PUT", "/test", nil)
	SignRequiredFunc(c)
	asserts.NotNil(c)
}

func TestWebDAVAuth(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()
	AuthFunc := WebDAVAuth()

	// options请求跳过验证
	{
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest("OPTIONS", "/test", nil)
		AuthFunc(c)
	}

	// 请求HTTP Basic Auth
	{
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest("POST", "/test", nil)
		AuthFunc(c)
		asserts.NotEmpty(c.Writer.Header()["WWW-Authenticate"])
	}

	// 用户名不存在
	{
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest("POST", "/test", nil)
		c.Request.Header = map[string][]string{
			"Authorization": {"Basic d2hvQGNsb3VkcmV2ZS5vcmc6YWRtaW4="},
		}
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "password", "email"}),
			)
		AuthFunc(c)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Equal(c.Writer.Status(), http.StatusUnauthorized)
	}

	// 密码错误
	{
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest("POST", "/test", nil)
		c.Request.Header = map[string][]string{
			"Authorization": {"Basic d2hvQGNsb3VkcmV2ZS5vcmc6YWRtaW4="},
		}
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "password", "email", "options"}).AddRow(1, "123", "who@cloudreve.org", "{}"),
			)
		AuthFunc(c)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Equal(c.Writer.Status(), http.StatusUnauthorized)
	}

	//未启用 WebDAV
	{
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest("POST", "/test", nil)
		c.Request.Header = map[string][]string{
			"Authorization": {"Basic d2hvQGNsb3VkcmV2ZS5vcmc6YWRtaW4="},
		}
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(
				sqlmock.NewRows(
					[]string{"id", "password", "email", "group_id", "options"}).
					AddRow(1,
						"rfBd67ti3SMtYvSg:ce6dc7bca4f17f2660e18e7608686673eae0fdf3",
						"who@cloudreve.org",
						1,
						"{}",
					),
			)
		mock.ExpectQuery("SELECT(.+)groups(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "web_dav_enabled"}).AddRow(1, false))
		AuthFunc(c)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Equal(c.Writer.Status(), http.StatusForbidden)
	}

	//正常
	{
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest("POST", "/test", nil)
		c.Request.Header = map[string][]string{
			"Authorization": {"Basic d2hvQGNsb3VkcmV2ZS5vcmc6YWRtaW4="},
		}
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(
				sqlmock.NewRows(
					[]string{"id", "password", "email", "group_id", "options"}).
					AddRow(1,
						"rfBd67ti3SMtYvSg:ce6dc7bca4f17f2660e18e7608686673eae0fdf3",
						"who@cloudreve.org",
						1,
						"{}",
					),
			)
		mock.ExpectQuery("SELECT(.+)groups(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "web_dav_enabled"}).AddRow(1, true))
		AuthFunc(c)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Equal(c.Writer.Status(), 200)
		_, ok := c.Get("user")
		asserts.True(ok)
	}

}

func TestRemoteCallbackAuth(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()
	AuthFunc := RemoteCallbackAuth()

	// 成功
	{
		cache.Set(
			"callback_testCallBackRemote",
			serializer.UploadSession{
				UID:         1,
				PolicyID:    2,
				VirtualPath: "/",
			},
			0,
		)
		cache.Deletes([]string{"1"}, "policy_")
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "group_id"}).AddRow(1, 1))
		mock.ExpectQuery("SELECT(.+)groups(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "policies"}).AddRow(1, "[2]"))
		mock.ExpectQuery("SELECT(.+)policies(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "secret_key"}).AddRow(2, "123"))
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"key", "testCallBackRemote"},
		}
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/remote/testCallBackRemote", nil)
		authInstance := auth.HMACAuth{SecretKey: []byte("123")}
		auth.SignRequest(authInstance, c.Request, 0)
		AuthFunc(c)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.False(c.IsAborted())
	}

	// Callback Key 不存在
	{

		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"key", "testCallBackRemote"},
		}
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/remote/testCallBackRemote", nil)
		authInstance := auth.HMACAuth{SecretKey: []byte("123")}
		auth.SignRequest(authInstance, c.Request, 0)
		AuthFunc(c)
		asserts.True(c.IsAborted())
	}

	// 用户不存在
	{
		cache.Set(
			"callback_testCallBackRemote",
			serializer.UploadSession{
				UID:         1,
				PolicyID:    2,
				VirtualPath: "/",
			},
			0,
		)
		cache.Deletes([]string{"1"}, "policy_")
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "group_id"}))
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"key", "testCallBackRemote"},
		}
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/remote/testCallBackRemote", nil)
		authInstance := auth.HMACAuth{SecretKey: []byte("123")}
		auth.SignRequest(authInstance, c.Request, 0)
		AuthFunc(c)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.True(c.IsAborted())
	}

	// 存储策略不一致
	{
		cache.Set(
			"callback_testCallBackRemote",
			serializer.UploadSession{
				UID:         1,
				PolicyID:    2,
				VirtualPath: "/",
			},
			0,
		)
		cache.Deletes([]string{"1"}, "policy_")
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "group_id"}).AddRow(1, 1))
		mock.ExpectQuery("SELECT(.+)groups(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "policies"}).AddRow(1, "[3]"))
		mock.ExpectQuery("SELECT(.+)policies(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "secret_key"}).AddRow(3, "123"))
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"key", "testCallBackRemote"},
		}
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/remote/testCallBackRemote", nil)
		authInstance := auth.HMACAuth{SecretKey: []byte("123")}
		auth.SignRequest(authInstance, c.Request, 0)
		AuthFunc(c)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.True(c.IsAborted())
	}

	// 签名错误
	{
		cache.Set(
			"callback_testCallBackRemote",
			serializer.UploadSession{
				UID:         1,
				PolicyID:    2,
				VirtualPath: "/",
			},
			0,
		)
		cache.Deletes([]string{"1"}, "policy_")
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "group_id"}).AddRow(1, 1))
		mock.ExpectQuery("SELECT(.+)groups(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "policies"}).AddRow(1, "[2]"))
		mock.ExpectQuery("SELECT(.+)policies(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "secret_key"}).AddRow(2, "123"))
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"key", "testCallBackRemote"},
		}
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/remote/testCallBackRemote", nil)
		AuthFunc(c)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.True(c.IsAborted())
	}

	// Callback Key 为空
	{
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/remote", nil)
		AuthFunc(c)
		asserts.True(c.IsAborted())
	}
}

func TestQiniuCallbackAuth(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()
	AuthFunc := QiniuCallbackAuth()

	// Callback Key 相关验证失败
	{
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"key", "testQiniuBackRemote"},
		}
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/remote/testQiniuBackRemote", nil)
		AuthFunc(c)
		asserts.True(c.IsAborted())
	}

	// 成功
	{
		cache.Set(
			"callback_testCallBackQiniu",
			serializer.UploadSession{
				UID:         1,
				PolicyID:    2,
				VirtualPath: "/",
			},
			0,
		)
		cache.Deletes([]string{"1"}, "policy_")
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "group_id"}).AddRow(1, 1))
		mock.ExpectQuery("SELECT(.+)groups(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "policies"}).AddRow(1, "[2]"))
		mock.ExpectQuery("SELECT(.+)policies(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "access_key", "secret_key"}).AddRow(2, "123", "123"))
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"key", "testCallBackQiniu"},
		}
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/qiniu/testCallBackQiniu", nil)
		mac := qbox.NewMac("123", "123")
		token, err := mac.SignRequest(c.Request)
		asserts.NoError(err)
		c.Request.Header["Authorization"] = []string{"QBox " + token}
		AuthFunc(c)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.False(c.IsAborted())
	}

	// 验证失败
	{
		cache.Set(
			"callback_testCallBackQiniu",
			serializer.UploadSession{
				UID:         1,
				PolicyID:    2,
				VirtualPath: "/",
			},
			0,
		)
		cache.Deletes([]string{"1"}, "policy_")
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "group_id"}).AddRow(1, 1))
		mock.ExpectQuery("SELECT(.+)groups(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "policies"}).AddRow(1, "[2]"))
		mock.ExpectQuery("SELECT(.+)policies(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "access_key", "secret_key"}).AddRow(2, "123", "123"))
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"key", "testCallBackQiniu"},
		}
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/qiniu/testCallBackQiniu", nil)
		mac := qbox.NewMac("123", "123")
		token, err := mac.SignRequest(c.Request)
		asserts.NoError(err)
		c.Request.Header["Authorization"] = []string{"QBox " + token + " "}
		AuthFunc(c)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.True(c.IsAborted())
	}
}

func TestOSSCallbackAuth(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()
	AuthFunc := OSSCallbackAuth()

	// Callback Key 相关验证失败
	{
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"key", "testOSSBackRemote"},
		}
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/oss/testQiniuBackRemote", nil)
		AuthFunc(c)
		asserts.True(c.IsAborted())
	}

	// 成功
	{
		cache.Set(
			"callback_testCallBackOSS",
			serializer.UploadSession{
				UID:         1,
				PolicyID:    2,
				VirtualPath: "/",
			},
			0,
		)
		cache.Deletes([]string{"1"}, "policy_")
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "group_id"}).AddRow(1, 1))
		mock.ExpectQuery("SELECT(.+)groups(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "policies"}).AddRow(1, "[2]"))
		mock.ExpectQuery("SELECT(.+)policies(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "access_key", "secret_key"}).AddRow(2, "123", "123"))
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"key", "testCallBackOSS"},
		}
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/qiniu/testCallBackOSS", nil)
		mac := qbox.NewMac("123", "123")
		token, err := mac.SignRequest(c.Request)
		asserts.NoError(err)
		c.Request.Header["Authorization"] = []string{"QBox " + token}
		AuthFunc(c)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.False(c.IsAborted())
	}

}
