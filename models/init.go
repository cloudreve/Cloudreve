package model

import (
	"Cloudreve/pkg/util"
	"github.com/jinzhu/gorm"
	"time"

	//
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// DB 数据库链接单例
var DB *gorm.DB

// Database 在中间件中初始化mysql链接
func Init() {
	//TODO 从配置文件中读取 包括DEBUG模式
	util.Log().Info("初始化数据库连接\n")
	db, err := gorm.Open("mysql", "root:root@(localhost)/v3?charset=utf8&parseTime=True&loc=Local")
	db.LogMode(true)
	// Error
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

	migration()
}
