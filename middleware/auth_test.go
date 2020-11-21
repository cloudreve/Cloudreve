package middleware

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/qiniu/api.v7/v7/auth/qbox"
	"github.com/stretchr/testify/assert"
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
		// 查找密码
		mock.ExpectQuery("SELECT(.+)webdav(.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}))
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
		// 查找密码
		mock.ExpectQuery("SELECT(.+)webdav(.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
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
		// 查找密码
		mock.ExpectQuery("SELECT(.+)webdav(.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
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
				PolicyID:    513,
				VirtualPath: "/",
			},
			0,
		)
		cache.Deletes([]string{"1"}, "policy_")
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "group_id"}).AddRow(1, 1))
		mock.ExpectQuery("SELECT(.+)groups(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "policies"}).AddRow(1, "[513]"))
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
				PolicyID:    550,
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

	// 签名错误
	{
		cache.Set(
			"callback_testCallBackRemote",
			serializer.UploadSession{
				UID:         1,
				PolicyID:    514,
				VirtualPath: "/",
			},
			0,
		)
		cache.Deletes([]string{"1"}, "policy_")
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "group_id"}).AddRow(1, 1))
		mock.ExpectQuery("SELECT(.+)groups(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "policies"}).AddRow(1, "[514]"))
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
				PolicyID:    515,
				VirtualPath: "/",
			},
			0,
		)
		cache.Deletes([]string{"1"}, "policy_")
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "group_id"}).AddRow(1, 1))
		mock.ExpectQuery("SELECT(.+)groups(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "policies"}).AddRow(1, "[515]"))
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
				PolicyID:    516,
				VirtualPath: "/",
			},
			0,
		)
		cache.Deletes([]string{"1"}, "policy_")
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "group_id"}).AddRow(1, 1))
		mock.ExpectQuery("SELECT(.+)groups(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "policies"}).AddRow(1, "[516]"))
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

	// 签名验证失败
	{
		cache.Set(
			"callback_testCallBackOSS",
			serializer.UploadSession{
				UID:         1,
				PolicyID:    517,
				VirtualPath: "/",
			},
			0,
		)
		cache.Deletes([]string{"1"}, "policy_")
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "group_id"}).AddRow(1, 1))
		mock.ExpectQuery("SELECT(.+)groups(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "policies"}).AddRow(1, "[517]"))
		mock.ExpectQuery("SELECT(.+)policies(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "access_key", "secret_key"}).AddRow(2, "123", "123"))
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"key", "testCallBackOSS"},
		}
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/oss/testCallBackOSS", nil)
		mac := qbox.NewMac("123", "123")
		token, err := mac.SignRequest(c.Request)
		asserts.NoError(err)
		c.Request.Header["Authorization"] = []string{"QBox " + token}
		AuthFunc(c)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.True(c.IsAborted())
	}

	// 成功
	{
		cache.Set(
			"callback_TnXx5E5VyfJUyM1UdkdDu1rtnJ34EbmH",
			serializer.UploadSession{
				UID:         1,
				PolicyID:    518,
				VirtualPath: "/",
			},
			0,
		)
		cache.Deletes([]string{"1"}, "policy_")
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "group_id"}).AddRow(1, 1))
		mock.ExpectQuery("SELECT(.+)groups(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "policies"}).AddRow(1, "[518]"))
		mock.ExpectQuery("SELECT(.+)policies(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "access_key", "secret_key"}).AddRow(2, "123", "123"))
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"key", "TnXx5E5VyfJUyM1UdkdDu1rtnJ34EbmH"},
		}
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/oss/TnXx5E5VyfJUyM1UdkdDu1rtnJ34EbmH", ioutil.NopCloser(strings.NewReader(`{"name":"2f7b2ccf30e9270ea920f1ab8a4037a546a2f0d5.jpg","source_name":"1/1_hFRtDLgM_2f7b2ccf30e9270ea920f1ab8a4037a546a2f0d5.jpg","size":114020,"pic_info":"810,539"}`)))
		c.Request.Header["Authorization"] = []string{"e5LwzwTkP9AFAItT4YzvdJOHd0Y0wqTMWhsV/h5SG90JYGAmMd+8LQyj96R+9qUfJWjMt6suuUh7LaOryR87Dw=="}
		c.Request.Header["X-Oss-Pub-Key-Url"] = []string{"aHR0cHM6Ly9nb3NzcHVibGljLmFsaWNkbi5jb20vY2FsbGJhY2tfcHViX2tleV92MS5wZW0="}
		AuthFunc(c)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.False(c.IsAborted())
	}

}

type fakeRead string

func (r fakeRead) Read(p []byte) (int, error) {
	return 0, errors.New("error")
}

func TestUpyunCallbackAuth(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()
	AuthFunc := UpyunCallbackAuth()

	// Callback Key 相关验证失败
	{
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"key", "testUpyunBackRemote"},
		}
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/upyun/testUpyunBackRemote", nil)
		AuthFunc(c)
		asserts.True(c.IsAborted())
	}

	// 无法获取请求正文
	{
		cache.Set(
			"callback_testCallBackUpyun",
			serializer.UploadSession{
				UID:         1,
				PolicyID:    509,
				VirtualPath: "/",
			},
			0,
		)
		cache.Deletes([]string{"1"}, "policy_")
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "group_id"}).AddRow(1, 1))
		mock.ExpectQuery("SELECT(.+)groups(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "policies"}).AddRow(1, "[519]"))
		mock.ExpectQuery("SELECT(.+)policies(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "access_key", "secret_key"}).AddRow(2, "123", "123"))
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"key", "testCallBackUpyun"},
		}
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/upyun/testCallBackUpyun", ioutil.NopCloser(fakeRead("")))
		AuthFunc(c)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.True(c.IsAborted())
	}

	// 正文MD5不一致
	{
		cache.Set(
			"callback_testCallBackUpyun",
			serializer.UploadSession{
				UID:         1,
				PolicyID:    510,
				VirtualPath: "/",
			},
			0,
		)
		cache.Deletes([]string{"1"}, "policy_")
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "group_id"}).AddRow(1, 1))
		mock.ExpectQuery("SELECT(.+)groups(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "policies"}).AddRow(1, "[520]"))
		mock.ExpectQuery("SELECT(.+)policies(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "access_key", "secret_key"}).AddRow(2, "123", "123"))
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"key", "testCallBackUpyun"},
		}
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/upyun/testCallBackUpyun", ioutil.NopCloser(strings.NewReader("1")))
		c.Request.Header["Content-Md5"] = []string{"123"}
		AuthFunc(c)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.True(c.IsAborted())
	}

	// 签名不一致
	{
		cache.Set(
			"callback_testCallBackUpyun",
			serializer.UploadSession{
				UID:         1,
				PolicyID:    511,
				VirtualPath: "/",
			},
			0,
		)
		cache.Deletes([]string{"1"}, "policy_")
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "group_id"}).AddRow(1, 1))
		mock.ExpectQuery("SELECT(.+)groups(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "policies"}).AddRow(1, "[521]"))
		mock.ExpectQuery("SELECT(.+)policies(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "access_key", "secret_key"}).AddRow(2, "123", "123"))
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"key", "testCallBackUpyun"},
		}
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/upyun/testCallBackUpyun", ioutil.NopCloser(strings.NewReader("1")))
		c.Request.Header["Content-Md5"] = []string{"c4ca4238a0b923820dcc509a6f75849b"}
		AuthFunc(c)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.True(c.IsAborted())
	}

	// 成功
	{
		cache.Set(
			"callback_testCallBackUpyun",
			serializer.UploadSession{
				UID:         1,
				PolicyID:    512,
				VirtualPath: "/",
			},
			0,
		)
		cache.Deletes([]string{"1"}, "policy_")
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "group_id"}).AddRow(1, 1))
		mock.ExpectQuery("SELECT(.+)groups(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "policies"}).AddRow(1, "[522]"))
		mock.ExpectQuery("SELECT(.+)policies(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "access_key", "secret_key"}).AddRow(2, "123", "123"))
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"key", "testCallBackUpyun"},
		}
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/upyun/testCallBackUpyun", ioutil.NopCloser(strings.NewReader("1")))
		c.Request.Header["Content-Md5"] = []string{"c4ca4238a0b923820dcc509a6f75849b"}
		c.Request.Header["Authorization"] = []string{"UPYUN 123:GWueK9x493BKFFk5gmfdO2Mn6EM="}
		AuthFunc(c)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.False(c.IsAborted())
	}
}

func TestOneDriveCallbackAuth(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()
	AuthFunc := OneDriveCallbackAuth()

	// Callback Key 相关验证失败
	{
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"key", "testUpyunBackRemote"},
		}
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/upyun/testUpyunBackRemote", nil)
		AuthFunc(c)
		asserts.True(c.IsAborted())
	}

	// 成功
	{
		cache.Set(
			"callback_testCallBackUpyun",
			serializer.UploadSession{
				UID:         1,
				PolicyID:    512,
				VirtualPath: "/",
			},
			0,
		)
		cache.Deletes([]string{"1"}, "policy_")
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "group_id"}).AddRow(1, 1))
		mock.ExpectQuery("SELECT(.+)groups(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "policies"}).AddRow(1, "[657]"))
		mock.ExpectQuery("SELECT(.+)policies(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "access_key", "secret_key"}).AddRow(2, "123", "123"))
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"key", "testCallBackUpyun"},
		}
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/upyun/testCallBackUpyun", ioutil.NopCloser(strings.NewReader("1")))
		AuthFunc(c)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.False(c.IsAborted())
	}
}

func TestCOSCallbackAuth(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()
	AuthFunc := COSCallbackAuth()

	// Callback Key 相关验证失败
	{
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"key", "testUpyunBackRemote"},
		}
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/upyun/testUpyunBackRemote", nil)
		AuthFunc(c)
		asserts.True(c.IsAborted())
	}

	// 成功
	{
		cache.Set(
			"callback_testCallBackUpyun",
			serializer.UploadSession{
				UID:         1,
				PolicyID:    512,
				VirtualPath: "/",
			},
			0,
		)
		cache.Deletes([]string{"1"}, "policy_")
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "group_id"}).AddRow(1, 1))
		mock.ExpectQuery("SELECT(.+)groups(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "policies"}).AddRow(1, "[702]"))
		mock.ExpectQuery("SELECT(.+)policies(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "access_key", "secret_key"}).AddRow(2, "123", "123"))
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"key", "testCallBackUpyun"},
		}
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/upyun/testCallBackUpyun", ioutil.NopCloser(strings.NewReader("1")))
		AuthFunc(c)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.False(c.IsAborted())
	}
}

func TestIsAdmin(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()
	testFunc := IsAdmin()

	// 非管理员
	{
		c, _ := gin.CreateTestContext(rec)
		c.Set("user", &model.User{})
		testFunc(c)
		asserts.True(c.IsAborted())
	}

	// 是管理员
	{
		c, _ := gin.CreateTestContext(rec)
		user := &model.User{}
		user.Group.ID = 1
		c.Set("user", user)
		testFunc(c)
		asserts.False(c.IsAborted())
	}

	// 初始用户，非管理组
	{
		c, _ := gin.CreateTestContext(rec)
		user := &model.User{}
		user.Group.ID = 2
		user.ID = 1
		c.Set("user", user)
		testFunc(c)
		asserts.False(c.IsAborted())
	}
}

func TestS3CallbackAuth(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()
	AuthFunc := S3CallbackAuth()

	// Callback Key 相关验证失败
	{
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"key", "testUpyunBackRemote"},
		}
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/upyun/testUpyunBackRemote", nil)
		AuthFunc(c)
		asserts.True(c.IsAborted())
	}

	// 成功
	{
		cache.Set(
			"callback_testCallBackUpyun",
			serializer.UploadSession{
				UID:         1,
				PolicyID:    512,
				VirtualPath: "/",
			},
			0,
		)
		cache.Deletes([]string{"1"}, "policy_")
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "group_id"}).AddRow(1, 1))
		mock.ExpectQuery("SELECT(.+)groups(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "policies"}).AddRow(1, "[702]"))
		mock.ExpectQuery("SELECT(.+)policies(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "access_key", "secret_key"}).AddRow(2, "123", "123"))
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"key", "testCallBackUpyun"},
		}
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/upyun/testCallBackUpyun", ioutil.NopCloser(strings.NewReader("1")))
		AuthFunc(c)
		asserts.False(c.IsAborted())
	}
}
