package conf

import "github.com/mojocn/base64Captcha"

// RedisConfig Redis服务器配置
var RedisConfig = &redis{
	Network:  "tcp",
	Server:   "",
	Password: "",
	DB:       "0",
}

// DatabaseConfig 数据库配置
var DatabaseConfig = &database{
	Type:    "UNSET",
	Charset: "utf8",
	DBFile:  "cloudreve.db",
	Port:    3306,
}

// SystemConfig 系统公用配置
var SystemConfig = &system{
	Debug:  false,
	Mode:   "master",
	Listen: ":5212",
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

// CORSConfig 跨域配置
var CORSConfig = &cors{
	AllowOrigins:     []string{"UNSET"},
	AllowMethods:     []string{"PUT", "POST", "GET", "OPTIONS"},
	AllowHeaders:     []string{"Cookie", "X-Cr-Policy", "Authorization", "Content-Length", "Content-Type", "X-Cr-Path", "X-Cr-FileName"},
	AllowCredentials: false,
	ExposeHeaders:    nil,
}

// ThumbConfig 缩略图配置
var ThumbConfig = &thumb{
	MaxWidth:      400,
	MaxHeight:     300,
	FileSuffix:    "._thumb",
	MaxTaskCount:  -1,
	EncodeMethod:  "jpg",
	GCAfterGen:    false,
	EncodeQuality: 85,
}

// SlaveConfig 从机配置
var SlaveConfig = &slave{
	CallbackTimeout: 20,
	SignatureTTL:    60,
}

var SSLConfig = &ssl{
	Listen:   ":443",
	CertPath: "",
	KeyPath:  "",
}

var UnixConfig = &unix{
	Listen: "",
}
