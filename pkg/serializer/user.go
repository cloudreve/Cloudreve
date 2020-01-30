package serializer

import (
	"fmt"
	"github.com/HFO4/cloudreve/models"
)

// CheckLogin 检查登录
func CheckLogin() Response {
	return Response{
		Code: CodeCheckLogin,
		Msg:  "未登录",
	}
}

// User 用户序列化器
type User struct {
	ID             uint   `json:"id"`
	Email          string `json:"user_name"`
	Nickname       string `json:"nickname"`
	Status         int    `json:"status"`
	Avatar         string `json:"avatar"`
	CreatedAt      int64  `json:"created_at"`
	PreferredTheme string `json:"preferred_theme"`
	Score          int    `json:"score"`
	Policy         policy `json:"policy"`
	Group          group  `json:"group"`
}

type policy struct {
	SaveType       string   `json:"saveType"`
	MaxSize        string   `json:"maxSize"`
	AllowedType    []string `json:"allowedType"`
	UploadURL      string   `json:"upUrl"`
	AllowGetSource bool     `json:"allowSource"`
}

type group struct {
	ID                   uint   `json:"id"`
	Name                 string `json:"name"`
	AllowShare           bool   `json:"allowShare"`
	AllowRemoteDownload  bool   `json:"allowRemoteDownload"`
	AllowTorrentDownload bool   `json:"allowTorrentDownload"`
	AllowArchiveDownload bool   `json:"allowArchiveDownload"`
	ShareFreeEnabled     bool   `json:"shareFree"`
	ShareDownload        bool   `json:"shareDownload"`
}

type storage struct {
	Used  uint64 `json:"used"`
	Free  uint64 `json:"free"`
	Total uint64 `json:"total"`
}

// BuildUser 序列化用户
func BuildUser(user model.User) User {
	aria2Option := user.Group.GetAria2Option()
	return User{
		ID:             user.ID,
		Email:          user.Email,
		Nickname:       user.Nick,
		Status:         user.Status,
		Avatar:         user.Avatar,
		CreatedAt:      user.CreatedAt.Unix(),
		PreferredTheme: user.OptionsSerialized.PreferredTheme,
		Score:          user.Score,
		Policy: policy{
			SaveType:       user.Policy.Type,
			MaxSize:        fmt.Sprintf("%.2fmb", float64(user.Policy.MaxSize)/(1024*1024)),
			AllowedType:    user.Policy.OptionsSerialized.FileType,
			UploadURL:      user.Policy.GetUploadURL(),
			AllowGetSource: user.Policy.IsOriginLinkEnable,
		},
		Group: group{
			ID:                   user.GroupID,
			Name:                 user.Group.Name,
			AllowShare:           user.Group.ShareEnabled,
			AllowRemoteDownload:  aria2Option[0],
			AllowTorrentDownload: aria2Option[2],
			AllowArchiveDownload: user.Group.OptionsSerialized.ArchiveDownloadEnabled,
			ShareFreeEnabled:     user.Group.OptionsSerialized.ShareFreeEnabled,
			ShareDownload:        user.Group.OptionsSerialized.ShareDownloadEnabled,
		},
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
