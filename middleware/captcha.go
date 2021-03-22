package middleware

import (
	"bytes"
	"encoding/json"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/recaptcha"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	captcha "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/captcha/v20190722"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"io"
	"io/ioutil"
	"strconv"
	"time"
)

type req struct {
	CaptchaCode string `json:"captchaCode"`
	Ticket      string `json:"ticket"`
	Randstr     string `json:"randstr"`
}

// CaptchaRequired 验证请求签名
func CaptchaRequired(configName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 相关设定
		options := model.GetSettingByNames(configName,
			"captcha_type",
			"captcha_ReCaptchaSecret",
			"captcha_TCaptcha_SecretId",
			"captcha_TCaptcha_SecretKey",
			"captcha_TCaptcha_CaptchaAppId",
			"captcha_TCaptcha_AppSecretKey")
		// 检查验证码
		isCaptchaRequired := model.IsTrueVal(options[configName])

		if isCaptchaRequired {
			var service req
			bodyCopy := new(bytes.Buffer)
			_, err := io.Copy(bodyCopy, c.Request.Body)
			if err != nil {
				c.JSON(200, serializer.ParamErr("验证码错误", err))
				c.Abort()
				return
			}

			bodyData := bodyCopy.Bytes()
			err = json.Unmarshal(bodyData, &service)
			if err != nil {
				c.JSON(200, serializer.ParamErr("验证码错误", err))
				c.Abort()
				return
			}

			c.Request.Body = ioutil.NopCloser(bytes.NewReader(bodyData))
			switch options["captcha_type"] {
			case "normal":
				captchaID := util.GetSession(c, "captchaID")
				util.DeleteSession(c, "captchaID")
				if captchaID == nil || !base64Captcha.VerifyCaptcha(captchaID.(string), service.CaptchaCode) {
					c.JSON(200, serializer.ParamErr("验证码错误", nil))
					c.Abort()
					return
				}

				break
			case "recaptcha":
				reCAPTCHA, err := recaptcha.NewReCAPTCHA(options["captcha_ReCaptchaSecret"], recaptcha.V2, 10*time.Second)
				if err != nil {
					util.Log().Warning("reCAPTCHA 验证错误, %s", err)
					c.Abort()
					break
				}

				err = reCAPTCHA.Verify(service.CaptchaCode)
				if err != nil {
					util.Log().Warning("reCAPTCHA 验证错误, %s", err)
					c.JSON(200, serializer.ParamErr("验证失败，请刷新网页后再次验证", nil))
					c.Abort()
					return
				}

				break
			case "tcaptcha":
				credential := common.NewCredential(
					options["captcha_TCaptcha_SecretId"],
					options["captcha_TCaptcha_SecretKey"],
				)
				cpf := profile.NewClientProfile()
				cpf.HttpProfile.Endpoint = "captcha.tencentcloudapi.com"
				client, _ := captcha.NewClient(credential, "", cpf)
				request := captcha.NewDescribeCaptchaResultRequest()
				request.CaptchaType = common.Uint64Ptr(9)
				appid, _ := strconv.Atoi(options["captcha_TCaptcha_CaptchaAppId"])
				request.CaptchaAppId = common.Uint64Ptr(uint64(appid))
				request.AppSecretKey = common.StringPtr(options["captcha_TCaptcha_AppSecretKey"])
				request.Ticket = common.StringPtr(service.Ticket)
				request.Randstr = common.StringPtr(service.Randstr)
				request.UserIp = common.StringPtr(c.ClientIP())
				response, err := client.DescribeCaptchaResult(request)
				if err != nil {
					util.Log().Warning("TCaptcha 验证错误, %s", err)
					c.Abort()
					break
				}

				if *response.Response.CaptchaCode != int64(1) {
					c.JSON(200, serializer.ParamErr("验证失败，请刷新网页后再次验证", nil))
					c.Abort()
					return
				}

				break
			}
		}
		c.Next()
	}
}
