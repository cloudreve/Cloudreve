package middleware

import (
	"database/sql"
	"errors"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/mq"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
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
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	authInstance := auth.HMACAuth{SecretKey: []byte(util.RandStringRunes(256))}
	SignRequiredFunc := SignRequired(authInstance)

	// 鉴权失败
	SignRequiredFunc(c)
	asserts.NotNil(c)
	asserts.True(c.IsAborted())

	c, _ = gin.CreateTestContext(rec)
	c.Request, _ = http.NewRequest("PUT", "/test", nil)
	SignRequiredFunc(c)
	asserts.NotNil(c)
	asserts.True(c.IsAborted())

	// Sign verify success
	c, _ = gin.CreateTestContext(rec)
	c.Request, _ = http.NewRequest("PUT", "/test", nil)
	c.Request = auth.SignRequest(authInstance, c.Request, 0)
	SignRequiredFunc(c)
	asserts.NotNil(c)
	asserts.False(c.IsAborted())
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

func TestUseUploadSession(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()
	AuthFunc := UseUploadSession("local")

	// sessionID 为空
	{

		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{}
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/remote/sessionID", nil)
		authInstance := auth.HMACAuth{SecretKey: []byte("123")}
		auth.SignRequest(authInstance, c.Request, 0)
		AuthFunc(c)
		asserts.True(c.IsAborted())
	}

	// 成功
	{
		cache.Set(
			filesystem.UploadSessionCachePrefix+"testCallBackRemote",
			serializer.UploadSession{
				UID:         1,
				VirtualPath: "/",
				Policy:      model.Policy{Type: "local"},
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
			{"sessionID", "testCallBackRemote"},
		}
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/remote/testCallBackRemote", nil)
		authInstance := auth.HMACAuth{SecretKey: []byte("123")}
		auth.SignRequest(authInstance, c.Request, 0)
		AuthFunc(c)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.False(c.IsAborted())
	}
}

func TestUploadCallbackCheck(t *testing.T) {
	a := assert.New(t)
	rec := httptest.NewRecorder()

	// 上传会话不存在
	{
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"sessionID", "testSessionNotExist"},
		}
		res := uploadCallbackCheck(c, "local")
		a.Contains("上传会话不存在或已过期", res.Msg)
	}

	// 上传策略不一致
	{
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"sessionID", "testPolicyNotMatch"},
		}
		cache.Set(
			filesystem.UploadSessionCachePrefix+"testPolicyNotMatch",
			serializer.UploadSession{
				UID:         1,
				VirtualPath: "/",
				Policy:      model.Policy{Type: "remote"},
			},
			0,
		)
		res := uploadCallbackCheck(c, "local")
		a.Contains("Policy not supported", res.Msg)
	}

	// 用户不存在
	{
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"sessionID", "testUserNotExist"},
		}
		cache.Set(
			filesystem.UploadSessionCachePrefix+"testUserNotExist",
			serializer.UploadSession{
				UID:         313,
				VirtualPath: "/",
				Policy:      model.Policy{Type: "remote"},
			},
			0,
		)
		mock.ExpectQuery("SELECT(.+)users(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "group_id"}))
		res := uploadCallbackCheck(c, "remote")
		a.Contains("找不到用户", res.Msg)
		a.NoError(mock.ExpectationsWereMet())
		_, ok := cache.Get(filesystem.UploadSessionCachePrefix + "testUserNotExist")
		a.False(ok)
	}
}

func TestRemoteCallbackAuth(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()
	AuthFunc := RemoteCallbackAuth()

	// 成功
	{
		c, _ := gin.CreateTestContext(rec)
		c.Set(filesystem.UploadSessionCtx, &serializer.UploadSession{
			UID:         1,
			VirtualPath: "/",
			Policy:      model.Policy{SecretKey: "123"},
		})
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/remote/testCallBackRemote", nil)
		authInstance := auth.HMACAuth{SecretKey: []byte("123")}
		auth.SignRequest(authInstance, c.Request, 0)
		AuthFunc(c)
		asserts.False(c.IsAborted())
	}

	// 签名错误
	{
		c, _ := gin.CreateTestContext(rec)
		c.Set(filesystem.UploadSessionCtx, &serializer.UploadSession{
			UID:         1,
			VirtualPath: "/",
			Policy:      model.Policy{SecretKey: "123"},
		})
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/remote/testCallBackRemote", nil)
		AuthFunc(c)
		asserts.True(c.IsAborted())
	}
}

func TestQiniuCallbackAuth(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()
	AuthFunc := QiniuCallbackAuth()

	// 成功
	{
		c, _ := gin.CreateTestContext(rec)
		c.Set(filesystem.UploadSessionCtx, &serializer.UploadSession{
			UID:         1,
			VirtualPath: "/",
			Policy: model.Policy{
				SecretKey: "123",
				AccessKey: "123",
			},
		})
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
		c, _ := gin.CreateTestContext(rec)
		c.Set(filesystem.UploadSessionCtx, &serializer.UploadSession{
			UID:         1,
			VirtualPath: "/",
			Policy: model.Policy{
				SecretKey: "123",
				AccessKey: "123",
			},
		})
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/qiniu/testCallBackQiniu", nil)
		mac := qbox.NewMac("123", "1213")
		token, err := mac.SignRequest(c.Request)
		asserts.NoError(err)
		c.Request.Header["Authorization"] = []string{"QBox " + token}
		AuthFunc(c)
		asserts.True(c.IsAborted())
	}
}

func TestOSSCallbackAuth(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()
	AuthFunc := OSSCallbackAuth()

	// 签名验证失败
	{
		c, _ := gin.CreateTestContext(rec)
		c.Set(filesystem.UploadSessionCtx, &serializer.UploadSession{
			UID:         1,
			VirtualPath: "/",
			Policy: model.Policy{
				SecretKey: "123",
				AccessKey: "123",
			},
		})
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/oss/testCallBackOSS", nil)
		mac := qbox.NewMac("123", "123")
		token, err := mac.SignRequest(c.Request)
		asserts.NoError(err)
		c.Request.Header["Authorization"] = []string{"QBox " + token}
		AuthFunc(c)
		asserts.True(c.IsAborted())
	}

	// 成功
	{
		c, _ := gin.CreateTestContext(rec)
		c.Set(filesystem.UploadSessionCtx, &serializer.UploadSession{
			UID:         1,
			VirtualPath: "/",
			Policy: model.Policy{
				SecretKey: "123",
				AccessKey: "123",
			},
		})
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/oss/TnXx5E5VyfJUyM1UdkdDu1rtnJ34EbmH", ioutil.NopCloser(strings.NewReader(`{"name":"2f7b2ccf30e9270ea920f1ab8a4037a546a2f0d5.jpg","source_name":"1/1_hFRtDLgM_2f7b2ccf30e9270ea920f1ab8a4037a546a2f0d5.jpg","size":114020,"pic_info":"810,539"}`)))
		c.Request.Header["Authorization"] = []string{"e5LwzwTkP9AFAItT4YzvdJOHd0Y0wqTMWhsV/h5SG90JYGAmMd+8LQyj96R+9qUfJWjMt6suuUh7LaOryR87Dw=="}
		c.Request.Header["X-Oss-Pub-Key-Url"] = []string{"aHR0cHM6Ly9nb3NzcHVibGljLmFsaWNkbi5jb20vY2FsbGJhY2tfcHViX2tleV92MS5wZW0="}
		AuthFunc(c)
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

	// 无法获取请求正文
	{
		c, _ := gin.CreateTestContext(rec)
		c.Set(filesystem.UploadSessionCtx, &serializer.UploadSession{
			UID:         1,
			VirtualPath: "/",
			Policy: model.Policy{
				SecretKey: "123",
				AccessKey: "123",
			},
		})
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/upyun/testCallBackUpyun", ioutil.NopCloser(fakeRead("")))
		AuthFunc(c)
		asserts.True(c.IsAborted())
	}

	// 正文MD5不一致
	{
		c, _ := gin.CreateTestContext(rec)
		c.Set(filesystem.UploadSessionCtx, &serializer.UploadSession{
			UID:         1,
			VirtualPath: "/",
			Policy: model.Policy{
				SecretKey: "123",
				AccessKey: "123",
			},
		})
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/upyun/testCallBackUpyun", ioutil.NopCloser(strings.NewReader("1")))
		c.Request.Header["Content-Md5"] = []string{"123"}
		AuthFunc(c)
		asserts.True(c.IsAborted())
	}

	// 签名不一致
	{
		c, _ := gin.CreateTestContext(rec)
		c.Set(filesystem.UploadSessionCtx, &serializer.UploadSession{
			UID:         1,
			VirtualPath: "/",
			Policy: model.Policy{
				SecretKey: "123",
				AccessKey: "123",
			},
		})
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/upyun/testCallBackUpyun", ioutil.NopCloser(strings.NewReader("1")))
		c.Request.Header["Content-Md5"] = []string{"c4ca4238a0b923820dcc509a6f75849b"}
		AuthFunc(c)
		asserts.True(c.IsAborted())
	}

	// 成功
	{
		c, _ := gin.CreateTestContext(rec)
		c.Set(filesystem.UploadSessionCtx, &serializer.UploadSession{
			UID:         1,
			VirtualPath: "/",
			Policy: model.Policy{
				SecretKey: "123",
				AccessKey: "123",
			},
		})
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/upyun/testCallBackUpyun", ioutil.NopCloser(strings.NewReader("1")))
		c.Request.Header["Content-Md5"] = []string{"c4ca4238a0b923820dcc509a6f75849b"}
		c.Request.Header["Authorization"] = []string{"UPYUN 123:GWueK9x493BKFFk5gmfdO2Mn6EM="}
		AuthFunc(c)
		asserts.False(c.IsAborted())
	}
}

func TestOneDriveCallbackAuth(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()
	AuthFunc := OneDriveCallbackAuth()

	// 成功
	{
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{
			{"sessionID", "TestOneDriveCallbackAuth"},
		}
		c.Set(filesystem.UploadSessionCtx, &serializer.UploadSession{
			UID:         1,
			VirtualPath: "/",
			Policy: model.Policy{
				SecretKey: "123",
				AccessKey: "123",
			},
		})
		c.Request, _ = http.NewRequest("POST", "/api/v3/callback/upyun/TestOneDriveCallbackAuth", ioutil.NopCloser(strings.NewReader("1")))
		res := mq.GlobalMQ.Subscribe("TestOneDriveCallbackAuth", 1)
		AuthFunc(c)
		select {
		case <-res:
		case <-time.After(time.Millisecond * 500):
			asserts.Fail("mq message should be published")
		}
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
