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
	Policy         Policy `json:"policy"`
	Group          Group  `json:"group"`
}

type Policy struct {
	SaveType    string   `json:"saveType"`
	MaxSize     string   `json:"maxSize"`
	AllowedType []string `json:"allowedType"`
	UploadURL   string   `json:"upUrl"`
}

type Group struct {
	AllowShare           bool `json:"allowShare"`
	AllowRemoteDownload  bool `json:"allowRemoteDownload"`
	AllowTorrentDownload bool `json:"allowTorrentDownload"`
}

// BuildUser 序列化用户
func BuildUser(user model.User) User {
	fmt.Println(user)
	return User{
		ID:             user.ID,
		Email:          user.Email,
		Nickname:       user.Nick,
		Status:         user.Status,
		Avatar:         user.Avatar,
		CreatedAt:      user.CreatedAt.Unix(),
		PreferredTheme: user.OptionsSerialized.PreferredTheme,
		Policy: Policy{
			SaveType:    user.Policy.Type,
			MaxSize:     fmt.Sprintf("%.2fmb", float64(user.Policy.MaxSize)/1024*1024),
			AllowedType: user.Policy.OptionsSerialized.FileType,
			UploadURL:   user.Policy.Server,
		},
		Group: Group{
			AllowShare:           user.Group.ShareEnabled,
			AllowRemoteDownload:  user.Group.Aria2Option[0] == '1',
			AllowTorrentDownload: user.Group.Aria2Option[2] == '1',
		},
	}
}

// BuildUserResponse 序列化用户响应
func BuildUserResponse(user model.User) Response {
	return Response{
		Data: BuildUser(user),
	}
}
