package conf

import (
	"cloudreve/pkg/util"
	"github.com/go-ini/ini"
	"github.com/mojocn/base64Captcha"
	"gopkg.in/go-playground/validator.v8"
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

// system 系统通用配置
type system struct {
	Debug         bool
	SessionSecret string
}

// captcha 验证码配置
type captcha struct {
	Height             int `validate:"gte=0"`
	Width              int `validate:"gte=0"`
	Mode               int `validate:"gte=0,lte=3"`
	ComplexOfNoiseText int `validate:"gte=0,lte=2"`
	ComplexOfNoiseDot  int `validate:"gte=0,lte=2"`
	IsShowHollowLine   bool
	IsShowNoiseDot     bool
	IsShowNoiseText    bool
	IsShowSlimeLine    bool
	IsShowSineLine     bool
	CaptchaLen         int `validate:"gte=0"`
}

// DatabaseConfig 数据库配置
var DatabaseConfig = &database{
	Type: "UNSET",
}

// SystemConfig 系统公用配置
var SystemConfig = &system{
	Debug: false,
}

// CaptchaConfig 验证码配置
var CaptchaConfig = &captcha{
	Height:             60,
	Width:              240,
	Mode:               3,
	ComplexOfNoiseText: base64Captcha.CaptchaComplexLower,
	ComplexOfNoiseDot:  base64Captcha.CaptchaComplexLower,
	IsShowHollowLine:   false,
	IsShowNoiseDot:     false,
	IsShowNoiseText:    false,
	IsShowSlimeLine:    false,
	IsShowSineLine:     false,
	CaptchaLen:         6,
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

	sections := map[string]interface{}{
		"Database": DatabaseConfig,
		"System":   SystemConfig,
		"Captcha":  CaptchaConfig,
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

	// 验证合法性
	validate := validator.New(&validator.Config{TagName: "validate"})
	err = validate.Struct(confStruct)
	if err != nil {
		return err
	}

	return nil
}
