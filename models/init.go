package model

import (
	"Cloudreve/pkg/conf"
	"Cloudreve/pkg/util"
	"fmt"
	"github.com/jinzhu/gorm"
	"time"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// DB 数据库链接单例
var DB *gorm.DB

// Database 在中间件中初始化mysql链接
func Init() {
	//TODO 从配置文件中读取 包括DEBUG模式
	util.Log().Info("初始化数据库连接\n")

	var (
		db  *gorm.DB
		err error
	)
	if conf.DatabaseConfig.Type == "UNSET" {
		//TODO 使用内置SQLite数据库
	} else {
		db, err = gorm.Open(conf.DatabaseConfig.Type, fmt.Sprintf("%s:%s@(%s)/%s?charset=utf8&parseTime=True&loc=Local",
			conf.DatabaseConfig.User,
			conf.DatabaseConfig.Password,
			conf.DatabaseConfig.Host,
			conf.DatabaseConfig.Name))
	}

	// 处理表前缀
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return conf.DatabaseConfig.TablePrefix + defaultTableName
	}

	db.LogMode(true)
	//db.SetLogger(util.Log())
	if err != nil {
		util.Log().Panic("连接数据库不成功", err)
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
