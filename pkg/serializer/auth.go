package serializer

import "encoding/json"

// RequestRawSign 待签名的HTTP请求
type RequestRawSign struct {
	Path   string
	Header string
	Body   string
}

// NewRequestSignString 返回JSON格式的待签名字符串
func NewRequestSignString(path, header, body string) string {
	req := RequestRawSign{
		Path:   path,
		Header: header,
		Body:   body,
	}
	res, _ := json.Marshal(req)
	return string(res)
}
