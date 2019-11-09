package conf

import (
	"Cloudreve/pkg/util"
	"github.com/go-ini/ini"
)

// Database 数据库
type database struct {
	Type        string
	User        string
	Password    string
	Host        string
	Name        string
	TablePrefix string
}

var DatabaseConfig = &database{
	Type: "UNSET",
}

var cfg *ini.File

// Init 初始化配置文件
func Init(path string) {
	var err error
	//TODO 配置文件不存在时创建
	//TODO 配置合法性验证
	cfg, err = ini.Load(path)
	if err != nil {
		util.Log().Panic("无法解析配置文件 '%s': ", path, err)
	}
	err = mapSection("Database", DatabaseConfig)
	if err != nil {
		util.Log().Warning("配置文件 %s 分区解析失败: ", "Database", err)
	}

}

// mapSection 将配置文件的 Section 映射到结构体上
func mapSection(section string, confStruct interface{}) error {
	err := cfg.Section(section).MapTo(confStruct)
	if err != nil {
		return err
	}
	return nil
}
