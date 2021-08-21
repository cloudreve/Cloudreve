package serializer

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
)

// Response 基础序列化器
type Response struct {
	Code  int         `json:"code"`
	Data  interface{} `json:"data,omitempty"`
	Msg   string      `json:"msg"`
	Error string      `json:"error,omitempty"`
}

// NewResponseWithGobData 返回Data字段使用gob编码的Response
func NewResponseWithGobData(data interface{}) Response {
	var w bytes.Buffer
	encoder := gob.NewEncoder(&w)
	if err := encoder.Encode(data); err != nil {
		return Err(CodeInternalSetting, "无法编码返回结果", err)
	}

	return Response{Data: w.Bytes()}
}

// GobDecode 将 Response 正文解码至目标指针
func (r *Response) GobDecode(target interface{}) {
	src := r.Data.(string)
	raw := make([]byte, len(src)*len(src)/base64.StdEncoding.DecodedLen(len(src)))
	base64.StdEncoding.Decode(raw, []byte(src))
	decoder := gob.NewDecoder(bytes.NewBuffer(raw))
	decoder.Decode(target)
}
