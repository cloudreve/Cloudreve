package aria2

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
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

func TestDummyAria2(t *testing.T) {
	asserts := assert.New(t)
	instance := DummyAria2{}
	asserts.Error(instance.CreateTask(nil, nil))
	_, err := instance.Status(nil)
	asserts.Error(err)
	asserts.Error(instance.Cancel(nil))
	asserts.Error(instance.Select(nil, nil))
}

func TestInit(t *testing.T) {
	MAX_RETRY = 0
	asserts := assert.New(t)
	cache.Set("setting_aria2_token", "1", 0)
	cache.Set("setting_aria2_call_timeout", "5", 0)
	cache.Set("setting_aria2_options", `[]`, 0)

	// 未指定RPC地址，跳过
	{
		cache.Set("setting_aria2_rpcurl", "", 0)
		Init(false)
		asserts.IsType(&DummyAria2{}, Instance)
	}

	// 无法解析服务器地址
	{
		cache.Set("setting_aria2_rpcurl", string(byte(0x7f)), 0)
		Init(false)
		asserts.IsType(&DummyAria2{}, Instance)
	}

	// 无法解析全局配置
	{
		Instance = &RPCService{}
		cache.Set("setting_aria2_options", "?", 0)
		cache.Set("setting_aria2_rpcurl", "ws://127.0.0.1:1234", 0)
		Init(false)
		asserts.IsType(&DummyAria2{}, Instance)
	}

	// 连接失败
	{
		cache.Set("setting_aria2_options", "{}", 0)
		cache.Set("setting_aria2_rpcurl", "http://127.0.0.1:1234", 0)
		cache.Set("setting_aria2_call_timeout", "1", 0)
		cache.Set("setting_aria2_interval", "100", 0)
		mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"g_id"}).AddRow("1"))
		Init(false)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.IsType(&RPCService{}, Instance)
	}
}

func TestGetStatus(t *testing.T) {
	asserts := assert.New(t)
	asserts.Equal(4, getStatus("complete"))
	asserts.Equal(1, getStatus("active"))
	asserts.Equal(0, getStatus("waiting"))
	asserts.Equal(2, getStatus("paused"))
	asserts.Equal(3, getStatus("error"))
	asserts.Equal(5, getStatus("removed"))
	asserts.Equal(6, getStatus("?"))
}
