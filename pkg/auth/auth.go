package auth

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/serializer"
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

// Init 初始化通用鉴权器
func Init() {
	General = HMACAuth{
		SecretKey: []byte(model.GetSettingByName("secret_key")),
	}
}
