package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"strconv"
	"strings"
	"time"
)

// HMACAuth HMAC算法鉴权
type HMACAuth struct {
	SecretKey []byte
}

// Sign 对给定Body生成expires后失效的签名，expires为过期时间戳，
// 填写为0表示不限制有效期
func (auth HMACAuth) Sign(body string, expires int64) string {
	h := hmac.New(sha256.New, auth.SecretKey)
	expireTimeStamp := strconv.FormatInt(expires, 10)
	_, err := io.WriteString(h, body+":"+expireTimeStamp)
	if err != nil {
		return ""
	}

	return base64.URLEncoding.EncodeToString(h.Sum(nil)) + ":" + expireTimeStamp
}

// Check 对给定Body和Sign进行鉴权，包括对expires的检查
func (auth HMACAuth) Check(body string, sign string) error {
	signSlice := strings.Split(sign, ":")
	// 如果未携带expires字段
	if signSlice[len(signSlice)-1] == "" {
		return ErrExpiresMissing
	}

	// 验证是否过期
	expires, err := strconv.ParseInt(signSlice[len(signSlice)-1], 10, 64)
	if err != nil {
		return ErrAuthFailed.WithError(err)
	}
	// 如果签名过期
	if expires < time.Now().Unix() && expires != 0 {
		return ErrExpired
	}

	// 验证签名
	if auth.Sign(body, expires) != sign {
		return ErrAuthFailed
	}
	return nil
}
