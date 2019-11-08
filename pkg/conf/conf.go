package conf

import (
	"Cloudreve/pkg/util"
	"fmt"
	"github.com/go-ini/ini"
)

type Conf struct {
	Database Database
}

type Database struct {
	Type        string
	User        string
	Password    string
	Host        string
	Name        string
	TablePrefix string
}

var database = &Database{
	Type: "UNSET",
}

var cfg *ini.File

func Init() {
	var err error
	//TODO 配置文件不存在时创建
	cfg, err = ini.Load("conf/conf.ini")
	if err != nil {
		util.Log().Panic("无法解析配置文件 'conf/conf.ini': ", err)
	}
	mapSection("Database", database)
	fmt.Println(database)

}

func mapSection(section string, confStruct interface{}) {
	err := cfg.Section("Database").MapTo(database)
	if err != nil {
		util.Log().Warning("配置文件 Database 分区解析失败")
	}
}
