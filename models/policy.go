package model

import (
	"encoding/gob"
	"encoding/json"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"

	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
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

// thumbSuffix 支持缩略图处理的文件扩展名
var thumbSuffix = map[string][]string{
	"local":    {},
	"qiniu":    {".psd", ".jpg", ".jpeg", ".png", ".gif", ".webp", ".tiff", ".bmp"},
	"oss":      {".jpg", ".jpeg", ".png", ".gif", ".webp", ".tiff", ".bmp"},
	"cos":      {".jpg", ".jpeg", ".png", ".gif", ".webp", ".tiff", ".bmp"},
	"upyun":    {".svg", ".jpg", ".jpeg", ".png", ".gif", ".webp", ".tiff", ".bmp"},
	"s3":       {},
	"remote":   {},
	"onedrive": {"*"},
}

func init() {
	// 注册缓存用到的复杂结构
	gob.Register(Policy{})
}

// GetPolicyByID 用ID获取存储策略
func GetPolicyByID(ID interface{}) (Policy, error) {
	// 尝试读取缓存
	cacheKey := "policy_" + strconv.Itoa(int(ID.(uint)))
	if policy, ok := cache.Get(cacheKey); ok {
		return policy.(Policy), nil
	}

	var policy Policy
	result := DB.First(&policy, ID)

	// 写入缓存
	if result.Error == nil {
		_ = cache.Set(cacheKey, policy, -1)
	}

	return policy, result.Error
}

// AfterFind 找到存储策略后的钩子
func (policy *Policy) AfterFind() (err error) {
	// 解析存储策略设置到OptionsSerialized
	if policy.Options != "" {
		err = json.Unmarshal([]byte(policy.Options), &policy.OptionsSerialized)
	}
	if policy.OptionsSerialized.FileType == nil {
		policy.OptionsSerialized.FileType = []string{}
	}

	return err
}

// BeforeSave Save策略前的钩子
func (policy *Policy) BeforeSave() (err error) {
	err = policy.SerializeOptions()
	return err
}

// SerializeOptions 将序列后的Option写入到数据库字段
func (policy *Policy) SerializeOptions() (err error) {
	optionsValue, err := json.Marshal(&policy.OptionsSerialized)
	policy.Options = string(optionsValue)
	return err
}

// GeneratePath 生成存储文件的路径
func (policy *Policy) GeneratePath(uid uint, origin string) string {
	dirRule := policy.DirNameRule
	replaceTable := map[string]string{
		"{randomkey16}":    util.RandStringRunes(16),
		"{randomkey8}":     util.RandStringRunes(8),
		"{timestamp}":      strconv.FormatInt(time.Now().Unix(), 10),
		"{timestamp_nano}": strconv.FormatInt(time.Now().UnixNano(), 10),
		"{uid}":            strconv.Itoa(int(uid)),
		"{datetime}":       time.Now().Format("20060102150405"),
		"{date}":           time.Now().Format("20060102"),
		"{year}":           time.Now().Format("2006"),
		"{month}":          time.Now().Format("01"),
		"{day}":            time.Now().Format("02"),
		"{hour}":           time.Now().Format("15"),
		"{minute}":         time.Now().Format("04"),
		"{second}":         time.Now().Format("05"),
		"{path}":           origin + "/",
	}
	dirRule = util.Replace(replaceTable, dirRule)
	return path.Clean(dirRule)
}

// GenerateFileName 生成存储文件名
func (policy *Policy) GenerateFileName(uid uint, origin string) string {
	// 未开启自动重命名时，直接返回原始文件名
	if !policy.AutoRename {
		return origin
	}

	fileRule := policy.FileNameRule

	replaceTable := map[string]string{
		"{randomkey16}":            util.RandStringRunes(16),
		"{randomkey8}":             util.RandStringRunes(8),
		"{timestamp}":              strconv.FormatInt(time.Now().Unix(), 10),
		"{timestamp_nano}":         strconv.FormatInt(time.Now().UnixNano(), 10),
		"{uid}":                    strconv.Itoa(int(uid)),
		"{datetime}":               time.Now().Format("20060102150405"),
		"{date}":                   time.Now().Format("20060102"),
		"{year}":                   time.Now().Format("2006"),
		"{month}":                  time.Now().Format("01"),
		"{day}":                    time.Now().Format("02"),
		"{hour}":                   time.Now().Format("15"),
		"{minute}":                 time.Now().Format("04"),
		"{second}":                 time.Now().Format("05"),
		"{originname}":             origin,
		"{ext}":                    filepath.Ext(origin),
		"{originname_without_ext}": strings.TrimSuffix(origin, filepath.Ext(origin)),
		"{uuid}":                   uuid.Must(uuid.NewV4()).String(),
	}

	fileRule = util.Replace(replaceTable, fileRule)
	return fileRule
}

// IsThumbExist 给定文件名，返回此存储策略下是否可能存在缩略图
func (policy *Policy) IsThumbExist(name string) bool {
	if list, ok := thumbSuffix[policy.Type]; ok {
		if len(list) == 1 && list[0] == "*" {
			return true
		}
		return util.ContainsString(list, strings.ToLower(filepath.Ext(name)))
	}
	return false
}

// IsDirectlyPreview 返回此策略下文件是否可以直接预览（不需要重定向）
func (policy *Policy) IsDirectlyPreview() bool {
	return policy.Type == "local"
}

// IsTransitUpload 返回此策略上传给定size文件时是否需要服务端中转
func (policy *Policy) IsTransitUpload(size uint64) bool {
	return policy.Type == "local"
}

// IsThumbGenerateNeeded 返回此策略是否需要在上传后生成缩略图
func (policy *Policy) IsThumbGenerateNeeded() bool {
	return policy.Type == "local"
}

// IsUploadPlaceholderWithSize 返回此策略创建上传会话时是否需要预留空间
func (policy *Policy) IsUploadPlaceholderWithSize() bool {
	if policy.Type == "remote" {
		return true
	}

	if util.ContainsString([]string{"onedrive", "oss", "qiniu", "cos", "s3"}, policy.Type) {
		return policy.OptionsSerialized.PlaceholderWithSize
	}

	return false
}

// CanStructureBeListed 返回存储策略是否能被前台列物理目录
func (policy *Policy) CanStructureBeListed() bool {
	return policy.Type != "local" && policy.Type != "remote"
}

// SaveAndClearCache 更新并清理缓存
func (policy *Policy) SaveAndClearCache() error {
	err := DB.Save(policy).Error
	policy.ClearCache()
	return err
}

// SaveAndClearCache 更新并清理缓存
func (policy *Policy) UpdateAccessKeyAndClearCache(s string) error {
	err := DB.Model(policy).UpdateColumn("access_key", s).Error
	policy.ClearCache()
	return err
}

// ClearCache 清空policy缓存
func (policy *Policy) ClearCache() {
	cache.Deletes([]string{strconv.FormatUint(uint64(policy.ID), 10)}, "policy_")
}

// CouldProxyThumb return if proxy thumbs is allowed for this policy.
func (policy *Policy) CouldProxyThumb() bool {
	if policy.Type == "local" || !IsTrueVal(GetSettingByName("thumb_proxy_enabled")) {
		return false
	}

	allowed := make([]uint, 0)
	_ = json.Unmarshal([]byte(GetSettingByName("thumb_proxy_policy")), &allowed)
	return lo.Contains[uint](allowed, policy.ID)
}
