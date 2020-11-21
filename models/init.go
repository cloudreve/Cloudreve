package model

import (
	"fmt"
	"time"

	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// DB 数据库链接单例
var DB *gorm.DB

// Init 初始化 MySQL 链接
func Init() {
	util.Log().Info("初始化数据库连接")

	var (
		db  *gorm.DB
		err error
	)

	if gin.Mode() == gin.TestMode {
		// 测试模式下，使用内存数据库
		db, err = gorm.Open("sqlite3", ":memory:")
	} else {
		switch conf.DatabaseConfig.Type {
		case "UNSET", "sqlite", "sqlite3":
			// 未指定数据库或者明确指定为 sqlite 时，使用 SQLite3 数据库
			db, err = gorm.Open("sqlite3", util.RelativePath(conf.DatabaseConfig.DBFile))
		case "mysql":
			// 当前只支持 sqlite3 与 mysql 数据库
			// TODO: import 其他 gorm 支持的主流数据库？否则直接 Open 没有任何意义。
			// TODO: 数据库连接其他参数允许用户自定义？譬如编码更换为 utf8mb4 以支持表情。
			db, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
				conf.DatabaseConfig.User,
				conf.DatabaseConfig.Password,
				conf.DatabaseConfig.Host,
				conf.DatabaseConfig.Port,
				conf.DatabaseConfig.Name))
		default:
			util.Log().Panic("不支持数据库类型: %s", conf.DatabaseConfig.Type)
		}
	}

	//db.SetLogger(util.Log())
	if err != nil {
		util.Log().Panic("连接数据库不成功, %s", err)
	}

	// 处理表前缀
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return conf.DatabaseConfig.TablePrefix + defaultTableName
	}

	// Debug模式下，输出所有 SQL 日志
	if conf.SystemConfig.Debug {
		db.LogMode(true)
	} else {
		db.LogMode(false)
	}

	//设置连接池
	//空闲
	db.DB().SetMaxIdleConns(50)
	//打开
	db.DB().SetMaxOpenConns(100)
	//超时
	db.DB().SetConnMaxLifetime(time.Second * 30)

	DB = db

	//执行迁移
	migration()
}
