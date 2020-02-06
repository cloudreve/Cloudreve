package model

import (
	"encoding/gob"
	"encoding/json"
	"github.com/HFO4/cloudreve/pkg/cache"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/jinzhu/gorm"
	"net/url"
	"path"
	"strconv"
	"time"
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
}

// PolicyOption 非公有的存储策略属性
type PolicyOption struct {
	// Upyun访问Token
	Token string `json:"token"`
	// 允许的文件扩展名
	FileType []string `json:"file_type"`
	// MimeType
	MimeType string `json:"mimetype"`

	// OdRedirect Onedrive重定向地址
	OdRedirect string `json:"od_redirect,omitempty"`
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
		if policy.OptionsSerialized.FileType == nil {
			policy.OptionsSerialized.FileType = []string{}
		}
	}

	return err
}

// BeforeSave Save策略前的钩子
func (policy *Policy) BeforeSave() (err error) {
	err = policy.SerializeOptions()
	return err
}

//SerializeOptions 将序列后的Option写入到数据库字段
func (policy *Policy) SerializeOptions() (err error) {
	optionsValue, err := json.Marshal(&policy.OptionsSerialized)
	policy.Options = string(optionsValue)
	return err
}

// GeneratePath 生成存储文件的路径
func (policy *Policy) GeneratePath(uid uint, origin string) string {
	dirRule := policy.DirNameRule
	replaceTable := map[string]string{
		"{randomkey16}": util.RandStringRunes(16),
		"{randomkey8}":  util.RandStringRunes(8),
		"{timestamp}":   strconv.FormatInt(time.Now().Unix(), 10),
		"{uid}":         strconv.Itoa(int(uid)),
		"{datetime}":    time.Now().Format("20060102150405"),
		"{date}":        time.Now().Format("20060102"),
		"{path}":        origin + "/",
	}
	dirRule = util.Replace(replaceTable, dirRule)
	return path.Clean(dirRule)
}

// GenerateFileName 生成存储文件名
func (policy *Policy) GenerateFileName(uid uint, origin string) string {
	// 未开启自动重命名时，直接返回原始文件名
	if !policy.AutoRename {
		return policy.getOriginNameRule(origin)
	}

	fileRule := policy.FileNameRule

	replaceTable := map[string]string{
		"{randomkey16}": util.RandStringRunes(16),
		"{randomkey8}":  util.RandStringRunes(8),
		"{timestamp}":   strconv.FormatInt(time.Now().Unix(), 10),
		"{uid}":         strconv.Itoa(int(uid)),
		"{datetime}":    time.Now().Format("20060102150405"),
		"{date}":        time.Now().Format("20060102"),
	}

	replaceTable["{originname}"] = policy.getOriginNameRule(origin)

	fileRule = util.Replace(replaceTable, fileRule)
	return fileRule
}

func (policy Policy) getOriginNameRule(origin string) string {
	// 部分存储策略可以使用{origin}代表原始文件名
	if origin == "" {
		// 如果上游未传回原始文件名，则使用占位符，让云存储端替换
		switch policy.Type {
		case "qiniu":
			// 七牛会将$(fname)自动替换为原始文件名
			return "$(fname)"
		case "local", "remote":
			return origin
		case "oss", "cos":
			// OSS会将${filename}自动替换为原始文件名
			return "${filename}"
		case "upyun":
			// Upyun会将{filename}{.suffix}自动替换为原始文件名
			return "{filename}{.suffix}"
		}
	}
	return origin
}

// IsDirectlyPreview 返回此策略下文件是否可以直接预览（不需要重定向）
func (policy *Policy) IsDirectlyPreview() bool {
	return policy.Type == "local"
}

// IsTransitUpload 返回此策略上传给定size文件时是否需要服务端中转
func (policy *Policy) IsTransitUpload(size uint64) bool {
	if policy.Type == "local" {
		return true
	}
	if policy.Type == "onedrive" && size < 4*1024*1024 {
		return true
	}
	return false
}

// IsPathGenerateNeeded 返回此策略是否需要在生成上传凭证时生成存储路径
func (policy *Policy) IsPathGenerateNeeded() bool {
	return policy.Type != "remote"
}

// IsThumbGenerateNeeded 返回此策略是否需要在上传后生成缩略图
func (policy *Policy) IsThumbGenerateNeeded() bool {
	return policy.Type == "local"
}

// IsMockThumbNeeded 返回此策略是否需要在上传后默认当图像文件
func (policy *Policy) IsMockThumbNeeded() bool {
	return policy.Type == "onedrive"
}

// GetUploadURL 获取文件上传服务API地址
func (policy *Policy) GetUploadURL() string {
	server, err := url.Parse(policy.Server)
	if err != nil {
		return policy.Server
	}

	var controller *url.URL
	switch policy.Type {
	case "local", "onedrive":
		return "/api/v3/file/upload"
	case "remote":
		controller, _ = url.Parse("/api/v3/slave/upload")
	case "oss", "cos":
		return policy.BaseURL
	case "upyun":
		return "http://v0.api.upyun.com/" + policy.BucketName
	default:
		controller, _ = url.Parse("")
	}
	return server.ResolveReference(controller).String()
}

// UpdateAccessKey 更新 AccessKey
func (policy *Policy) UpdateAccessKey(key string) error {
	policy.AccessKey = key
	err := DB.Save(policy).Error
	policy.ClearCache()
	return err
}

// ClearCache 清空policy缓存
func (policy *Policy) ClearCache() {
	cache.Deletes([]string{strconv.FormatUint(uint64(policy.ID), 10)}, "policy_")
}
