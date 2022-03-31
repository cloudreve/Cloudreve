package serializer

import (
	"encoding/gob"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"time"
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
	SessionID   string   `json:"sessionID"`
	ChunkSize   uint64   `json:"chunkSize"` // 分块大小，0 为部分快
	Expires     int64    `json:"expires"`   // 上传凭证过期时间， Unix 时间戳
	UploadURLs  []string `json:"uploadURLs,omitempty"`
	Credential  string   `json:"credential,omitempty"`
	UploadID    string   `json:"uploadID,omitempty"`
	Callback    string   `json:"callback,omitempty"` // 回调地址
	Path        string   `json:"path,omitempty"`     // 存储路径
	AccessKey   string   `json:"ak,omitempty"`
	KeyTime     string   `json:"keyTime,omitempty"` // COS用有效期
	Policy      string   `json:"policy,omitempty"`
	CompleteURL string   `json:"completeURL,omitempty"`
}

// UploadSession 上传会话
type UploadSession struct {
	Key            string     // 上传会话 GUID
	UID            uint       // 发起者
	VirtualPath    string     // 用户文件路径，不含文件名
	Name           string     // 文件名
	Size           uint64     // 文件大小
	SavePath       string     // 物理存储路径，包含物理文件名
	LastModified   *time.Time // 可选的文件最后修改日期
	Policy         model.Policy
	Callback       string // 回调 URL 地址
	CallbackSecret string // 回调 URL
	UploadURL      string
	UploadID       string
	Credential     string
}

// UploadCallback 上传回调正文
type UploadCallback struct {
	PicInfo string `json:"pic_info"`
}

// GeneralUploadCallbackFailed 存储策略上传回调失败响应
type GeneralUploadCallbackFailed struct {
	Error string `json:"error"`
}

func init() {
	gob.Register(UploadSession{})
}
