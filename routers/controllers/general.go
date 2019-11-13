package controllers

import (
	"cloudreve/pkg/serializer"
	"cloudreve/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
)

// Captcha 获取验证码
func Captcha(c *gin.Context) {

	var configD = base64Captcha.ConfigCharacter{
		Height: 60,
		Width:  240,
		//const CaptchaModeNumber:数字,CaptchaModeAlphabet:字母,CaptchaModeArithmetic:算术,CaptchaModeNumberAlphabet:数字字母混合.
		Mode:               base64Captcha.CaptchaModeNumberAlphabet,
		ComplexOfNoiseText: base64Captcha.CaptchaComplexLower,
		ComplexOfNoiseDot:  base64Captcha.CaptchaComplexLower,
		IsShowHollowLine:   false,
		IsShowNoiseDot:     false,
		IsShowNoiseText:    false,
		IsShowSlimeLine:    false,
		IsShowSineLine:     false,
		CaptchaLen:         6,
	}

	idKeyD, capD := base64Captcha.GenerateCaptcha("", configD)
	util.SetSession(c, map[string]interface{}{
		"captchaID": idKeyD,
	})
	base64stringD := base64Captcha.CaptchaWriteToBase64Encoding(capD)

	c.JSON(200, serializer.Response{
		Code: 0,
		Data: base64stringD,
	})
}
