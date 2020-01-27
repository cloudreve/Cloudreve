package serializer

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/hashid"
	"time"
)

// Share 分享序列化
type Share struct {
	Locked     bool          `json:"locked"`
	IsDir      bool          `json:"is_dir"`
	Score      int           `json:"score"`
	CreateDate string        `json:"create_date,omitempty"`
	Downloads  int           `json:"downloads"`
	Views      int           `json:"views"`
	Expire     int64         `json:"expire"`
	Creator    *shareCreator `json:"creator,omitempty"`
	Source     *shareSource  `json:"source,omitempty"`
}

type shareCreator struct {
	Key       string `json:"key"`
	Nick      string `json:"nick"`
	GroupName string `json:"group_name"`
}

type shareSource struct {
	Name string `json:"name"`
	Size uint64 `json:"size"`
}

// BuildShareResponse 构建获取分享信息响应
func BuildShareResponse(share *model.Share, unlocked bool) Share {
	creator := share.GetCreator()
	resp := Share{
		Locked: !unlocked,
		Creator: &shareCreator{
			Key:       hashid.HashID(creator.ID, hashid.UserID),
			Nick:      creator.Nick,
			GroupName: creator.Group.Name,
		},
		Score:      share.Score,
		CreateDate: share.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if !unlocked {
		return resp
	}

	resp.IsDir = share.IsDir
	resp.Downloads = share.Downloads
	resp.Views = share.Views
	if share.Expires != nil {
		resp.Expire = share.Expires.Unix() - time.Now().Unix()
	}

	if share.IsDir {

	} else {
		source := share.GetSourceFile()
		resp.Source = &shareSource{
			Name: source.Name,
			Size: source.Size,
		}
	}

	return resp

}
