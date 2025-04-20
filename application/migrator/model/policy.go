package model

import (
	"github.com/jinzhu/gorm"
)

// Policy 存储策略
type Policy struct {
	// 表字段
	gorm.Model
	Name               string
	Type               string
	Server             string
	BucketName         string
	IsPrivate          bool
	BaseURL            string
	AccessKey          string `gorm:"type:text"`
	SecretKey          string `gorm:"type:text"`
	MaxSize            uint64
	AutoRename         bool
	DirNameRule        string
	FileNameRule       string
	IsOriginLinkEnable bool
	Options            string `gorm:"type:text"`

	// 数据库忽略字段
	OptionsSerialized PolicyOption `gorm:"-"`
	MasterID          string       `gorm:"-"`
}

// PolicyOption 非公有的存储策略属性
type PolicyOption struct {
	// Upyun访问Token
	Token string `json:"token"`
	// 允许的文件扩展名
	FileType []string `json:"file_type"`
	// MimeType
	MimeType string `json:"mimetype"`
	// OauthRedirect Oauth 重定向地址
	OauthRedirect string `json:"od_redirect,omitempty"`
	// OdProxy Onedrive 反代地址
	OdProxy string `json:"od_proxy,omitempty"`
	// OdDriver OneDrive 驱动器定位符
	OdDriver string `json:"od_driver,omitempty"`
	// Region 区域代码
	Region string `json:"region,omitempty"`
	// ServerSideEndpoint 服务端请求使用的 Endpoint，为空时使用 Policy.Server 字段
	ServerSideEndpoint string `json:"server_side_endpoint,omitempty"`
	// 分片上传的分片大小
	ChunkSize uint64 `json:"chunk_size,omitempty"`
	// 分片上传时是否需要预留空间
	PlaceholderWithSize bool `json:"placeholder_with_size,omitempty"`
	// 每秒对存储端的 API 请求上限
	TPSLimit float64 `json:"tps_limit,omitempty"`
	// 每秒 API 请求爆发上限
	TPSLimitBurst int `json:"tps_limit_burst,omitempty"`
	// Set this to `true` to force the request to use path-style addressing,
	// i.e., `http://s3.amazonaws.com/BUCKET/KEY `
	S3ForcePathStyle bool `json:"s3_path_style"`
	// File extensions that support thumbnail generation using native policy API.
	ThumbExts []string `json:"thumb_exts,omitempty"`
}
