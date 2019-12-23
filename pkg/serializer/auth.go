package serializer

import "encoding/json"

// RequestRawSign 待签名的HTTP请求
type RequestRawSign struct {
	Path   string
	Policy string
	Body   string
}

// NewRequestSignString 返回JSON格式的待签名字符串
// TODO 测试
func NewRequestSignString(path, policy, body string) string {
	req := RequestRawSign{
		Path:   path,
		Policy: policy,
		Body:   body,
	}
	res, _ := json.Marshal(req)
	return string(res)
}
