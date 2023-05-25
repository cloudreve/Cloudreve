package googledrive

import "encoding/gob"

// RespError 接口返回错误
type RespError struct {
	APIError APIError `json:"error"`
}

// APIError 接口返回的错误内容
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error 实现error接口
func (err RespError) Error() string {
	return err.APIError.Message
}

// Credential 获取token时返回的凭证
type Credential struct {
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
}

// OAuthError OAuth相关接口的错误响应
type OAuthError struct {
	ErrorType        string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// Error 实现error接口
func (err OAuthError) Error() string {
	return err.ErrorDescription
}

func init() {
	gob.Register(Credential{})
}
