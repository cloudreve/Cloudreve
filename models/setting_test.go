package model

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"testing"
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
	asserts := assert.New(t)

	//找到设置时
	rows := sqlmock.NewRows([]string{"name", "value", "type"}).
		AddRow("siteName", "Cloudreve", "basic").
		AddRow("siteDes", "Something wonderful", "basic")
	mock.ExpectQuery("^SELECT \\* FROM `(.+)` WHERE `(.+)`\\.`deleted_at` IS NULL AND(.+)$").WillReturnRows(rows)
	settings := GetSettingByNames([]string{"siteName", "siteDes"})
	asserts.Equal(map[string]string{
		"siteName": "Cloudreve",
		"siteDes":  "Something wonderful",
	}, settings)
	asserts.NoError(mock.ExpectationsWereMet())

	//找到其中一个设置时
	rows = sqlmock.NewRows([]string{"name", "value", "type"}).
		AddRow("siteName", "Cloudreve", "basic")
	mock.ExpectQuery("^SELECT \\* FROM `(.+)` WHERE `(.+)`\\.`deleted_at` IS NULL AND(.+)$").WillReturnRows(rows)
	settings = GetSettingByNames([]string{"siteName", "siteDes2333"})
	asserts.Equal(map[string]string{
		"siteName": "Cloudreve",
	}, settings)
	asserts.NoError(mock.ExpectationsWereMet())

	//找不到设置时
	rows = sqlmock.NewRows([]string{"name", "value", "type"})
	mock.ExpectQuery("^SELECT \\* FROM `(.+)` WHERE `(.+)`\\.`deleted_at` IS NULL AND(.+)$").WillReturnRows(rows)
	settings = GetSettingByNames([]string{"siteName2333", "siteDes2333"})
	asserts.Equal(map[string]string{}, settings)
	asserts.NoError(mock.ExpectationsWereMet())
}

// TestGetSettingByName 测试GetSettingByName
func TestGetSettingByName(t *testing.T) {
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
