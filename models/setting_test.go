package model

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

var mock sqlmock.Sqlmock
var mockDB *gorm.DB

// TestMain 初始化数据库Mock
func TestMain(m *testing.M) {
	var db *sql.DB
	var err error
	db, mock, err = sqlmock.New()
	if err != nil {
		panic("An error was not expected when opening a stub database connection")
	}
	DB, _ = gorm.Open("mysql", db)
	mockDB = DB
	defer db.Close()
	m.Run()
}

func TestGetSettingByType(t *testing.T) {
	cache.Store = cache.NewMemoStore()
	asserts := assert.New(t)

	//找到设置时
	rows := sqlmock.NewRows([]string{"name", "value", "type"}).
		AddRow("siteName", "Cloudreve", "basic").
		AddRow("siteDes", "Something wonderful", "basic")
	mock.ExpectQuery("^SELECT \\* FROM `(.+)` WHERE `(.+)`\\.`deleted_at` IS NULL AND(.+)$").WillReturnRows(rows)
	settings := GetSettingByType([]string{"basic"})
	asserts.Equal(map[string]string{
		"siteName": "Cloudreve",
		"siteDes":  "Something wonderful",
	}, settings)

	rows = sqlmock.NewRows([]string{"name", "value", "type"}).
		AddRow("siteName", "Cloudreve", "basic").
		AddRow("siteDes", "Something wonderful", "basic2")
	mock.ExpectQuery("^SELECT \\* FROM `(.+)` WHERE `(.+)`\\.`deleted_at` IS NULL AND(.+)$").WillReturnRows(rows)
	settings = GetSettingByType([]string{"basic", "basic2"})
	asserts.Equal(map[string]string{
		"siteName": "Cloudreve",
		"siteDes":  "Something wonderful",
	}, settings)

	//找不到
	rows = sqlmock.NewRows([]string{"name", "value", "type"})
	mock.ExpectQuery("^SELECT \\* FROM `(.+)` WHERE `(.+)`\\.`deleted_at` IS NULL AND(.+)$").WillReturnRows(rows)
	settings = GetSettingByType([]string{"basic233"})
	asserts.Equal(map[string]string{}, settings)
}

func TestGetSettingByNames(t *testing.T) {
	cache.Store = cache.NewMemoStore()
	asserts := assert.New(t)

	//找到设置时
	rows := sqlmock.NewRows([]string{"name", "value", "type"}).
		AddRow("siteName", "Cloudreve", "basic").
		AddRow("siteDes", "Something wonderful", "basic")
	mock.ExpectQuery("^SELECT \\* FROM `(.+)` WHERE `(.+)`\\.`deleted_at` IS NULL AND(.+)$").WillReturnRows(rows)
	settings := GetSettingByNames("siteName", "siteDes")
	asserts.Equal(map[string]string{
		"siteName": "Cloudreve",
		"siteDes":  "Something wonderful",
	}, settings)
	asserts.NoError(mock.ExpectationsWereMet())

	//找到其中一个设置时
	rows = sqlmock.NewRows([]string{"name", "value", "type"}).
		AddRow("siteName2", "Cloudreve", "basic")
	mock.ExpectQuery("^SELECT \\* FROM `(.+)` WHERE `(.+)`\\.`deleted_at` IS NULL AND(.+)$").WillReturnRows(rows)
	settings = GetSettingByNames("siteName2", "siteDes2333")
	asserts.Equal(map[string]string{
		"siteName2": "Cloudreve",
	}, settings)
	asserts.NoError(mock.ExpectationsWereMet())

	//找不到设置时
	rows = sqlmock.NewRows([]string{"name", "value", "type"})
	mock.ExpectQuery("^SELECT \\* FROM `(.+)` WHERE `(.+)`\\.`deleted_at` IS NULL AND(.+)$").WillReturnRows(rows)
	settings = GetSettingByNames("siteName2333", "siteDes2333")
	asserts.Equal(map[string]string{}, settings)
	asserts.NoError(mock.ExpectationsWereMet())

	// 一个设置命中缓存
	mock.ExpectQuery("^SELECT \\* FROM `(.+)` WHERE `(.+)`\\.`deleted_at` IS NULL AND(.+)$").WithArgs("siteDes2").WillReturnRows(sqlmock.NewRows([]string{"name", "value", "type"}).
		AddRow("siteDes2", "Cloudreve2", "basic"))
	settings = GetSettingByNames("siteName", "siteDes2")
	asserts.Equal(map[string]string{
		"siteName": "Cloudreve",
		"siteDes2": "Cloudreve2",
	}, settings)
	asserts.NoError(mock.ExpectationsWereMet())

}

// TestGetSettingByName 测试GetSettingByName
func TestGetSettingByName(t *testing.T) {
	cache.Store = cache.NewMemoStore()
	asserts := assert.New(t)

	//找到设置时
	rows := sqlmock.NewRows([]string{"name", "value", "type"}).
		AddRow("siteName", "Cloudreve", "basic")
	mock.ExpectQuery("^SELECT \\* FROM `(.+)` WHERE `(.+)`\\.`deleted_at` IS NULL AND(.+)$").WillReturnRows(rows)

	siteName := GetSettingByName("siteName")
	asserts.Equal("Cloudreve", siteName)
	asserts.NoError(mock.ExpectationsWereMet())

	// 第二次查询应返回缓存内容
	siteNameCache := GetSettingByName("siteName")
	asserts.Equal("Cloudreve", siteNameCache)
	asserts.NoError(mock.ExpectationsWereMet())

	// 找不到设置
	rows = sqlmock.NewRows([]string{"name", "value", "type"})
	mock.ExpectQuery("^SELECT \\* FROM `(.+)` WHERE `(.+)`\\.`deleted_at` IS NULL AND(.+)$").WillReturnRows(rows)

	siteName = GetSettingByName("siteName not exist")
	asserts.Equal("", siteName)
	asserts.NoError(mock.ExpectationsWereMet())

}

func TestIsTrueVal(t *testing.T) {
	asserts := assert.New(t)

	asserts.True(IsTrueVal("1"))
	asserts.True(IsTrueVal("true"))
	asserts.False(IsTrueVal("0"))
	asserts.False(IsTrueVal("false"))
}

func TestGetSiteURL(t *testing.T) {
	asserts := assert.New(t)

	// 正常
	{
		err := cache.Deletes([]string{"siteURL"}, "setting_")
		asserts.NoError(err)

		mock.ExpectQuery("SELECT(.+)").WithArgs("siteURL").WillReturnRows(sqlmock.NewRows([]string{"id", "value"}).AddRow(1, "https://drive.cloudreve.org"))
		siteURL := GetSiteURL()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Equal("https://drive.cloudreve.org", siteURL.String())
	}

	// 失败 返回默认值
	{
		err := cache.Deletes([]string{"siteURL"}, "setting_")
		asserts.NoError(err)

		mock.ExpectQuery("SELECT(.+)").WithArgs("siteURL").WillReturnRows(sqlmock.NewRows([]string{"id", "value"}).AddRow(1, ":][\\/\\]sdf"))
		siteURL := GetSiteURL()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Equal("https://cloudreve.org", siteURL.String())
	}
}

func TestGetIntSetting(t *testing.T) {
	asserts := assert.New(t)

	// 正常
	{
		cache.Set("setting_TestGetIntSetting", "10", 0)
		res := GetIntSetting("TestGetIntSetting", 20)
		asserts.Equal(10, res)
	}

	// 使用默认值
	{
		res := GetIntSetting("TestGetIntSetting_2", 20)
		asserts.Equal(20, res)
	}

}
