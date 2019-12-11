package auth

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/serializer"
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
// TODO slave模式下从配置文件获取
func Init() {
	General = HMACAuth{
		SecretKey: []byte(model.GetSettingByName("secret_key")),
	}
}
