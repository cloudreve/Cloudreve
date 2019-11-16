package controllers

import (
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
)

// Captcha 获取验证码
func Captcha(c *gin.Context) {
	// 验证码配置
	var configD = base64Captcha.ConfigCharacter{
		Height: conf.CaptchaConfig.Height,
		Width:  conf.CaptchaConfig.Width,
		//const CaptchaModeNumber:数字,CaptchaModeAlphabet:字母,CaptchaModeArithmetic:算术,CaptchaModeNumberAlphabet:数字字母混合.
		Mode:               conf.CaptchaConfig.Mode,
		ComplexOfNoiseText: conf.CaptchaConfig.ComplexOfNoiseText,
		ComplexOfNoiseDot:  conf.CaptchaConfig.ComplexOfNoiseDot,
		IsShowHollowLine:   conf.CaptchaConfig.IsShowHollowLine,
		IsShowNoiseDot:     conf.CaptchaConfig.IsShowNoiseDot,
		IsShowNoiseText:    conf.CaptchaConfig.IsShowNoiseText,
		IsShowSlimeLine:    conf.CaptchaConfig.IsShowSlimeLine,
		IsShowSineLine:     conf.CaptchaConfig.IsShowSineLine,
		CaptchaLen:         conf.CaptchaConfig.CaptchaLen,
	}

	// 生成验证码
	idKeyD, capD := base64Captcha.GenerateCaptcha("", configD)
	// 将验证码UID存入Session以便后续验证
	util.SetSession(c, map[string]interface{}{
		"captchaID": idKeyD,
	})

	// 将验证码图像编码为Base64
	base64stringD := base64Captcha.CaptchaWriteToBase64Encoding(capD)

	c.JSON(200, serializer.Response{
		Code: 0,
		Data: base64stringD,
	})
}
