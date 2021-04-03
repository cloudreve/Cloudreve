package serializer

import (
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/hashid"
)

// Share 分享信息序列化
type Share struct {
	Key        string        `json:"key"`
	Locked     bool          `json:"locked"`
	IsDir      bool          `json:"is_dir"`
	CreateDate time.Time     `json:"create_date,omitempty"`
	Downloads  int           `json:"downloads"`
	Views      int           `json:"views"`
	Expire     int64         `json:"expire"`
	Preview    bool          `json:"preview"`
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

// myShareItem 我的分享列表条目
type myShareItem struct {
	Key             string       `json:"key"`
	IsDir           bool         `json:"is_dir"`
	Password        string       `json:"password"`
	CreateDate      time.Time    `json:"create_date,omitempty"`
	Downloads       int          `json:"downloads"`
	RemainDownloads int          `json:"remain_downloads"`
	Views           int          `json:"views"`
	Expire          int64        `json:"expire"`
	Preview         bool         `json:"preview"`
	Source          *shareSource `json:"source,omitempty"`
}

// BuildShareList 构建我的分享列表响应
func BuildShareList(shares []model.Share, total int) Response {
	res := make([]myShareItem, 0, total)
	now := time.Now().Unix()
	for i := 0; i < len(shares); i++ {
		item := myShareItem{
			Key:             hashid.HashID(shares[i].ID, hashid.ShareID),
			IsDir:           shares[i].IsDir,
			Password:        shares[i].Password,
			CreateDate:      shares[i].CreatedAt,
			Downloads:       shares[i].Downloads,
			Views:           shares[i].Views,
			Preview:         shares[i].PreviewEnabled,
			Expire:          -1,
			RemainDownloads: shares[i].RemainDownloads,
		}
		if shares[i].Expires != nil {
			item.Expire = shares[i].Expires.Unix() - now
			if item.Expire == 0 {
				item.Expire = 0
			}
		}
		if shares[i].File.ID != 0 {
			item.Source = &shareSource{
				Name: shares[i].File.Name,
				Size: shares[i].File.Size,
			}
		} else if shares[i].Folder.ID != 0 {
			item.Source = &shareSource{
				Name: shares[i].Folder.Name,
			}
		}

		res = append(res, item)
	}

	return Response{Data: map[string]interface{}{
		"total": total,
		"items": res,
	}}
}

// BuildShareResponse 构建获取分享信息响应
func BuildShareResponse(share *model.Share, unlocked bool) Share {
	creator := share.Creator()
	resp := Share{
		Key:    hashid.HashID(share.ID, hashid.ShareID),
		Locked: !unlocked,
		Creator: &shareCreator{
			Key:       hashid.HashID(creator.ID, hashid.UserID),
			Nick:      creator.Nick,
			GroupName: creator.Group.Name,
		},
		CreateDate: share.CreatedAt,
	}

	// 未解锁时只返回基本信息
	if !unlocked {
		return resp
	}

	resp.IsDir = share.IsDir
	resp.Downloads = share.Downloads
	resp.Views = share.Views
	resp.Preview = share.PreviewEnabled

	if share.Expires != nil {
		resp.Expire = share.Expires.Unix() - time.Now().Unix()
	}

	if share.IsDir {
		source := share.SourceFolder()
		resp.Source = &shareSource{
			Name: source.Name,
			Size: 0,
		}
	} else {
		source := share.SourceFile()
		resp.Source = &shareSource{
			Name: source.Name,
			Size: source.Size,
		}
	}

	return resp

}
