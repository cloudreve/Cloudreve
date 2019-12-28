package serializer

import (
	"encoding/base64"
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
	CallbackKey      string   `json:"callback_key"`
}

// DecodeUploadPolicy 反序列化Header中携带的上传策略
// TODO 测试
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
