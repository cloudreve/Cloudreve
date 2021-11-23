package auth

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

var (
	ErrAuthFailed        = serializer.NewError(serializer.CodeNoPermissionErr, "鉴权失败", nil)
	ErrAuthHeaderMissing = serializer.NewError(serializer.CodeNoPermissionErr, "authorization header is missing", nil)
	ErrExpiresMissing    = serializer.NewError(serializer.CodeNoPermissionErr, "expire timestamp is missing", nil)
	ErrExpired           = serializer.NewError(serializer.CodeSignExpired, "签名已过期", nil)
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

// SignRequest 对PUT\POST等复杂HTTP请求签名，只会对URI部分、
// 请求正文、`X-Cr-`开头的header进行签名
func SignRequest(instance Auth, r *http.Request, expires int64) *http.Request {
	// 处理有效期
	if expires > 0 {
		expires += time.Now().Unix()
	}

	// 生成签名
	sign := instance.Sign(getSignContent(r), expires)

	// 将签名加到请求Header中
	r.Header["Authorization"] = []string{"Bearer " + sign}
	return r
}

// CheckRequest 对复杂请求进行签名验证
func CheckRequest(instance Auth, r *http.Request) error {
	var (
		sign []string
		ok   bool
	)
	if sign, ok = r.Header["Authorization"]; !ok || len(sign) == 0 {
		return ErrAuthHeaderMissing
	}
	sign[0] = strings.TrimPrefix(sign[0], "Bearer ")

	return instance.Check(getSignContent(r), sign[0])
}

// getSignContent 签名请求 path、正文、以`X-`开头的 Header. 如果 Header 中包含 `X-Policy`，
// 则不对正文签名。返回待签名/验证的字符串
func getSignContent(r *http.Request) (rawSignString string) {
	// 读取所有body正文
	var body = []byte{}
	if _, ok := r.Header["X-Cr-Policy"]; !ok {
		if r.Body != nil {
			body, _ = ioutil.ReadAll(r.Body)
			_ = r.Body.Close()
			r.Body = ioutil.NopCloser(bytes.NewReader(body))
		}
	}

	// 决定要签名的header
	var signedHeader []string
	for k, _ := range r.Header {
		if strings.HasPrefix(k, "X-Cr-") && k != "X-Cr-Filename" {
			signedHeader = append(signedHeader, fmt.Sprintf("%s=%s", k, r.Header.Get(k)))
		}
	}
	sort.Strings(signedHeader)

	// 读取所有待签名Header
	rawSignString = serializer.NewRequestSignString(r.URL.Path, strings.Join(signedHeader, "&"), string(body))

	return rawSignString
}

// SignURI 对URI进行签名,签名只针对Path部分，query部分不做验证
func SignURI(instance Auth, uri string, expires int64) (*url.URL, error) {
	// 处理有效期
	if expires != 0 {
		expires += time.Now().Unix()
	}

	base, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	// 生成签名
	sign := instance.Sign(base.Path, expires)

	// 将签名加到URI中
	queries := base.Query()
	queries.Set("sign", sign)
	base.RawQuery = queries.Encode()

	return base, nil
}

// CheckURI 对URI进行鉴权
func CheckURI(instance Auth, url *url.URL) error {
	//获取待验证的签名正文
	queries := url.Query()
	sign := queries.Get("sign")
	queries.Del("sign")
	url.RawQuery = queries.Encode()

	return instance.Check(url.Path, sign)
}

// Init 初始化通用鉴权器
func Init() {
	var secretKey string
	if conf.SystemConfig.Mode == "master" {
		secretKey = model.GetSettingByName("secret_key")
	} else {
		secretKey = conf.SlaveConfig.Secret
		if secretKey == "" {
			util.Log().Panic("未指定 SlaveSecret，请前往配置文件中指定")
		}
	}
	General = HMACAuth{
		SecretKey: []byte(secretKey),
	}
}
