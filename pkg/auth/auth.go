package auth

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/pkg/util"
	"io/ioutil"
	"net/http"
	"net/url"
)

var (
	ErrAuthFailed = serializer.NewError(serializer.CodeNoRightErr, "鉴权失败", nil)
	ErrExpired    = serializer.NewError(serializer.CodeSignExpired, "签名已过期", nil)
)

// General 通用的认证接口
var General Auth

// Auth 鉴权认证
type Auth interface {
	// 对给定Body进行签名,expires为0表示永不过期
	Sign(body string, expires int64) string
	// 对给定Body和Sign进行检查
	Check(body string, sign string) error
}

// SignRequest 对PUT\POST等复杂HTTP请求签名，如果请求Header中
// 包含 X-Policy， 则此请求会被认定为上传请求，只会对URI部分和
// Policy部分进行签名。其他请求则会对URI和Body部分进行签名。
func SignRequest(r *http.Request, expires int64) *http.Request {
	var rawSignString string
	if policy, ok := r.Header["X-Policy"]; ok {
		rawSignString = serializer.NewRequestSignString(r.URL.Path, policy[0], "")
	} else {
		body, _ := ioutil.ReadAll(r.Body)
		rawSignString = serializer.NewRequestSignString(r.URL.Path, "", string(body))
	}

	// 生成签名
	sign := General.Sign(rawSignString, expires)

	// 将签名加到请求Header中
	r.Header["Authorization"] = []string{"Bearer " + sign}
	return r
}

// SignURI 对URI进行签名,签名只针对Path部分，query部分不做验证
func SignURI(uri string, expires int64) (*url.URL, error) {
	base, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	// 生成签名
	sign := General.Sign(base.Path, expires)

	// 将签名加到URI中
	queries := base.Query()
	queries.Set("sign", sign)
	base.RawQuery = queries.Encode()

	return base, nil
}

// CheckURI 对URI进行鉴权
func CheckURI(url *url.URL) error {
	//获取待验证的签名正文
	queries := url.Query()
	sign := queries.Get("sign")
	queries.Del("sign")
	url.RawQuery = queries.Encode()

	return General.Check(url.Path, sign)
}

// Init 初始化通用鉴权器
// TODO 测试
func Init() {
	var secretKey string
	if conf.SystemConfig.Mode == "master" {
		secretKey = model.GetSettingByName("secret_key")
	} else {
		secretKey = conf.SystemConfig.SlaveSecret
		if secretKey == "" {
			util.Log().Panic("未指定 SlaveSecret，请前往配置文件中指定")
		}
	}
	General = HMACAuth{
		SecretKey: []byte(secretKey),
	}
}
