package serializer

import (
	"encoding/gob"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/hashid"
	"time"
)

func init() {
	gob.Register(ObjectProps{})
}

// ObjectProps 文件、目录对象的详细属性信息
type ObjectProps struct {
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Policy         string    `json:"policy"`
	Size           uint64    `json:"size"`
	ChildFolderNum int       `json:"child_folder_num"`
	ChildFileNum   int       `json:"child_file_num"`
	Path           string    `json:"path"`

	QueryDate time.Time `json:"query_date"`
}

// ObjectList 文件、目录列表
type ObjectList struct {
	Parent  string         `json:"parent,omitempty"`
	Objects []Object       `json:"objects"`
	Policy  *PolicySummary `json:"policy,omitempty"`
}

// Object 文件或者目录
type Object struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Path          string    `json:"path"`
	Thumb         bool      `json:"thumb"`
	Size          uint64    `json:"size"`
	Type          string    `json:"type"`
	Date          time.Time `json:"date"`
	CreateDate    time.Time `json:"create_date"`
	Key           string    `json:"key,omitempty"`
	SourceEnabled bool      `json:"source_enabled"`
}

// PolicySummary 用于前端组件使用的存储策略概况
type PolicySummary struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	MaxSize  uint64   `json:"max_size"`
	FileType []string `json:"file_type"`
}

// BuildObjectList 构建列目录响应
func BuildObjectList(parent uint, objects []Object, policy *model.Policy) ObjectList {
	res := ObjectList{
		Objects: objects,
	}

	if parent > 0 {
		res.Parent = hashid.HashID(parent, hashid.FolderID)
	}

	if policy != nil {
		res.Policy = &PolicySummary{
			ID:       hashid.HashID(policy.ID, hashid.PolicyID),
			Name:     policy.Name,
			Type:     policy.Type,
			MaxSize:  policy.MaxSize,
			FileType: policy.OptionsSerialized.FileType,
		}
	}

	return res
}

// Sources 获取外链的结果响应
type Sources struct {
	URL    string `json:"url"`
	Name   string `json:"name"`
	Parent uint   `json:"parent"`
	Error  string `json:"error,omitempty"`
}

// DocPreviewSession 文档预览会话响应
type DocPreviewSession struct {
	URL            string `json:"url"`
	AccessToken    string `json:"access_token,omitempty"`
	AccessTokenTTL int64  `json:"access_token_ttl,omitempty"`
}

// WopiFileInfo Response for `CheckFileInfo`
type WopiFileInfo struct {
	// Required
	BaseFileName string
	Version      string
	Size         int64

	// Breadcrumb
	BreadcrumbBrandName  string
	BreadcrumbBrandUrl   string
	BreadcrumbFolderName string
	BreadcrumbFolderUrl  string

	// Post Message
	FileSharingPostMessage bool
	ClosePostMessage       bool
	PostMessageOrigin      string

	// Other miscellaneous properties
	FileNameMaxLength int
	LastModifiedTime  string

	// User metadata
	IsAnonymousUser  bool
	UserFriendlyName string
	UserId           string
	OwnerId          string

	// Permission
	ReadOnly      bool
	UserCanRename bool
	UserCanReview bool
	UserCanWrite  bool

	SupportsRename    bool
	SupportsReviewing bool
	SupportsUpdate    bool
}
