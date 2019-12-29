package conf

import (
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/go-ini/ini"
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
	Mode          string `validate:"eq=master|eq=slave"`
	Listen        string `validate:"required"`
	Debug         bool
	SessionSecret string
}

// slave 作为slave存储端配置
type slave struct {
	Secret          string `validate:"omitempty,gte=64"`
	CallbackTimeout int    `validate:"omitempty,gte=1"`
	SignatureTTL    int    `validate:"omitempty,gte=1"`
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
	CaptchaLen         int `validate:"gt=0"`
}

// redis 配置
type redis struct {
	Server   string
	Password string
	DB       string
}

// 缩略图 配置
type thumb struct {
	MaxWidth   uint
	MaxHeight  uint
	FileSuffix string `validate:"min=1"`
}

// 跨域配置
type cors struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	ExposeHeaders    []string
}

var cfg *ini.File

// Init 初始化配置文件
func Init(path string) {
	var err error
	//TODO 配置文件不存在时创建
	//TODO 配置合法性验证
	cfg, err = ini.Load(path)
	if err != nil {
		util.Log().Panic("无法解析配置文件 '%s': %s", path, err)
	}

	sections := map[string]interface{}{
		"Database":  DatabaseConfig,
		"System":    SystemConfig,
		"Captcha":   CaptchaConfig,
		"Redis":     RedisConfig,
		"Thumbnail": ThumbConfig,
		"CORS":      CORSConfig,
		"Slave":     SlaveConfig,
	}
	for sectionName, sectionStruct := range sections {
		err = mapSection(sectionName, sectionStruct)
		if err != nil {
			util.Log().Panic("配置文件 %s 分区解析失败: %s", sectionName, err)
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
