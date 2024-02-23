package serializer

import (
	"fmt"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/hashid"
	"github.com/duo-labs/webauthn/webauthn"
)

// CheckLogin 检查登录
func CheckLogin() Response {
	return Response{
		Code: CodeCheckLogin,
		Msg:  "Login required",
	}
}

// PhoneRequired 需要绑定手机
func PhoneRequired() Response {
	return Response{
		Code: CodePhoneRequired,
		Msg:  "此功能需要绑定手机后使用",
	}
}

// User 用户序列化器
type User struct {
	ID             string    `json:"id"`
	Email          string    `json:"user_name"`
	Nickname       string    `json:"nickname"`
	Status         int       `json:"status"`
	Avatar         string    `json:"avatar"`
	CreatedAt      time.Time `json:"created_at"`
	PreferredTheme string    `json:"preferred_theme"`
	Score          int       `json:"score"`
	Anonymous      bool      `json:"anonymous"`
	Group          group     `json:"group"`
	Tags           []tag     `json:"tags"`
}

type group struct {
	ID                   uint   `json:"id"`
	Name                 string `json:"name"`
	AllowShare           bool   `json:"allowShare"`
	AllowRemoteDownload  bool   `json:"allowRemoteDownload"`
	AllowArchiveDownload bool   `json:"allowArchiveDownload"`
	ShareFreeEnabled     bool   `json:"shareFree"`
	ShareDownload        bool   `json:"shareDownload"`
	CompressEnabled      bool   `json:"compress"`
	WebDAVEnabled        bool   `json:"webdav"`
	RelocateEnabled      bool   `json:"relocate"`
	SourceBatchSize      int    `json:"sourceBatch"`
	SelectNode           bool   `json:"selectNode"`
	AdvanceDelete        bool   `json:"advanceDelete"`
	AllowWebDAVProxy     bool   `json:"allowWebDAVProxy"`
}

type tag struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Icon       string `json:"icon"`
	Color      string `json:"color"`
	Type       int    `json:"type"`
	Expression string `json:"expression"`
}

type storage struct {
	Used  uint64 `json:"used"`
	Free  uint64 `json:"free"`
	Total uint64 `json:"total"`
}

// WebAuthnCredentials 外部验证器凭证
type WebAuthnCredentials struct {
	ID          []byte `json:"id"`
	FingerPrint string `json:"fingerprint"`
}

// BuildWebAuthnList 构建设置页面凭证列表
func BuildWebAuthnList(credentials []webauthn.Credential) []WebAuthnCredentials {
	res := make([]WebAuthnCredentials, 0, len(credentials))
	for _, v := range credentials {
		credential := WebAuthnCredentials{
			ID:          v.ID,
			FingerPrint: fmt.Sprintf("% X", v.Authenticator.AAGUID),
		}
		res = append(res, credential)
	}

	return res
}

// BuildUser 序列化用户
func BuildUser(user model.User) User {
	tags, _ := model.GetTagsByUID(user.ID)
	return User{
		ID:             hashid.HashID(user.ID, hashid.UserID),
		Email:          user.Email,
		Nickname:       user.Nick,
		Status:         user.Status,
		Avatar:         user.Avatar,
		CreatedAt:      user.CreatedAt,
		PreferredTheme: user.OptionsSerialized.PreferredTheme,
		Score:          user.Score,
		Anonymous:      user.IsAnonymous(),
		Group: group{
			ID:                   user.GroupID,
			Name:                 user.Group.Name,
			AllowShare:           user.Group.ShareEnabled,
			AllowRemoteDownload:  user.Group.OptionsSerialized.Aria2,
			AllowArchiveDownload: user.Group.OptionsSerialized.ArchiveDownload,
			ShareFreeEnabled:     user.Group.OptionsSerialized.ShareFree,
			ShareDownload:        user.Group.OptionsSerialized.ShareDownload,
			CompressEnabled:      user.Group.OptionsSerialized.ArchiveTask,
			WebDAVEnabled:        user.Group.WebDAVEnabled,
			AllowWebDAVProxy:     user.Group.OptionsSerialized.WebDAVProxy,
			RelocateEnabled:      user.Group.OptionsSerialized.Relocate,
			SourceBatchSize:      user.Group.OptionsSerialized.SourceBatchSize,
			SelectNode:           user.Group.OptionsSerialized.SelectNode,
			AdvanceDelete:        user.Group.OptionsSerialized.AdvanceDelete,
		},
		Tags: buildTagRes(tags),
	}
}

// BuildUserResponse 序列化用户响应
func BuildUserResponse(user model.User) Response {
	return Response{
		Data: BuildUser(user),
	}
}

// BuildUserStorageResponse 序列化用户存储概况响应
func BuildUserStorageResponse(user model.User) Response {
	total := user.Group.MaxStorage + user.GetAvailablePackSize()
	storageResp := storage{
		Used:  user.Storage,
		Free:  total - user.Storage,
		Total: total,
	}

	if total < user.Storage {
		storageResp.Free = 0
	}

	return Response{
		Data: storageResp,
	}
}

// buildTagRes 构建标签列表
func buildTagRes(tags []model.Tag) []tag {
	res := make([]tag, 0, len(tags))
	for i := 0; i < len(tags); i++ {
		newTag := tag{
			ID:    hashid.HashID(tags[i].ID, hashid.TagID),
			Name:  tags[i].Name,
			Icon:  tags[i].Icon,
			Color: tags[i].Color,
			Type:  tags[i].Type,
		}
		if newTag.Type != 0 {
			newTag.Expression = tags[i].Expression

		}
		res = append(res, newTag)
	}

	return res
}
