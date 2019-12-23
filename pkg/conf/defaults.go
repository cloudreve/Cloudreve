package conf

import "github.com/mojocn/base64Captcha"

// RedisConfig Redis服务器配置
var RedisConfig = &redis{
	Server:   "",
	Password: "",
	DB:       "0",
}

// DatabaseConfig 数据库配置
var DatabaseConfig = &database{
	Type: "UNSET",
}

// SystemConfig 系统公用配置
var SystemConfig = &system{
	Debug:  false,
	Mode:   "master",
	Listen: ":5000",
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
	AllowAllOrigins:  false,
	AllowOrigins:     []string{"UNSET"},
	AllowMethods:     []string{"PUT", "POST", "GET", "OPTIONS"},
	AllowHeaders:     []string{"Cookie", "Content-Length", "Content-Type", "X-Path", "X-FileName"},
	AllowCredentials: true,
	ExposeHeaders:    nil,
}

var ThumbConfig = &thumb{
	MaxWidth:   400,
	MaxHeight:  300,
	FileSuffix: "._thumb",
}
