package serializer

import (
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
)

// UploadPolicy slave模式下传递的上传策略
type UploadPolicy struct {
	SavePath         string   `json:"save_path"`
	FileName         string   `json:"file_name"`
	AutoRename       bool     `json:"auto_rename"`
	MaxSize          uint64   `json:"max_size"`
	AllowedExtension []string `json:"allowed_extension"`
	CallbackURL      string   `json:"callback_url"`
}

// UploadCredential 返回给客户端的上传凭证
type UploadCredential struct {
	Token     string `json:"token"`
	Policy    string `json:"policy"`
	Path      string `json:"path"` // 存储路径
	AccessKey string `json:"ak"`
	KeyTime   string `json:"key_time,omitempty"` // COS用有效期
	Callback  string `json:"callback,omitempty"` // 回调地址
	Key       string `json:"key,omitempty"`      // 文件标识符，通常为回调key
}

// UploadSession 上传会话
type UploadSession struct {
	Key         string
	UID         uint
	PolicyID    uint
	VirtualPath string
	Name        string
	Size        uint64
	SavePath    string
}

// UploadCallback 上传回调正文
type UploadCallback struct {
	Name       string `json:"name"`
	SourceName string `json:"source_name"`
	PicInfo    string `json:"pic_info"`
	Size       uint64 `json:"size"`
}

// GeneralUploadCallbackFailed 存储策略上传回调失败响应
type GeneralUploadCallbackFailed struct {
	Error string `json:"error"`
}

func init() {
	gob.Register(UploadSession{})
}

// DecodeUploadPolicy 反序列化Header中携带的上传策略
func DecodeUploadPolicy(raw string) (*UploadPolicy, error) {
	var res UploadPolicy

	rawJSON, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(rawJSON, &res)
	if err != nil {
		return nil, err
	}

	return &res, err
}

// EncodeUploadPolicy 序列化Header中携带的上传策略
func (policy *UploadPolicy) EncodeUploadPolicy() (string, error) {
	jsonRes, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}

	res := base64.StdEncoding.EncodeToString(jsonRes)
	return res, nil

}
