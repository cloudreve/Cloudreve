package middleware

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/gin-gonic/gin"
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
