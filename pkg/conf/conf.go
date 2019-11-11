package conf

import (
	"cloudreve/pkg/util"
	"github.com/go-ini/ini"
)

// database 数据库
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

// system 系统通用配置
type system struct {
	Debug         bool
	SessionSecret string
}

var SystemConfig = &system{}

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

	sections := map[string]interface{}{
		"Database": DatabaseConfig,
		"System":   SystemConfig,
	}
	for sectionName, sectionStruct := range sections {
		err = mapSection(sectionName, sectionStruct)
		if err != nil {
			util.Log().Warning("配置文件 %s 分区解析失败: ", sectionName, err)
		}
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
