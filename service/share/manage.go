package share

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/hashid"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/gin-gonic/gin"
	"net/url"
	"time"
)

// ShareCreateService 创建新分享服务
type ShareCreateService struct {
	SourceID        uint   `json:"id" binding:"required"`
	IsDir           bool   `json:"is_dir"`
	Password        string `json:"password" binding:"max=255"`
	RemainDownloads int    `json:"downloads"`
	Expire          int    `json:"expire"`
	Score           int    `json:"score" binding:"gte=0"`
}

// Create 创建新分享
func (service *ShareCreateService) Create(c *gin.Context) serializer.Response {
	userCtx, _ := c.Get("user")
	user := userCtx.(*model.User)

	// 是否拥有权限
	if !user.Group.ShareEnabled {
		return serializer.Err(serializer.CodeNoRightErr, "您无权创建分享链接", nil)
	}

	// 对象是否存在
	exist := true
	if service.IsDir {
		folder, err := model.GetFoldersByIDs([]uint{service.SourceID}, user.ID)
		if err != nil || len(folder) == 0 {
			exist = false
		}
	} else {
		file, err := model.GetFilesByIDs([]uint{service.SourceID}, user.ID)
		if err != nil || len(file) == 0 {
			exist = false
		}
	}
	if !exist {
		return serializer.Err(serializer.CodeNotFound, "原始资源不存在", nil)
	}

	newShare := model.Share{
		Password: service.Password,
		IsDir:    service.IsDir,
		UserID:   user.ID,
		SourceID: service.SourceID,
		Score:    service.RemainDownloads,
	}

	// 如果开启了自动过期
	if service.RemainDownloads > 0 {
		expires := time.Now().Add(time.Duration(service.Expire) * time.Second)
		newShare.RemainDownloads = service.RemainDownloads
		newShare.Expires = &expires
	}

	// 创建分享
	id, err := newShare.Create()
	if err != nil {
		return serializer.Err(serializer.CodeDBError, "分享链接创建失败", err)
	}

	// 获取分享的唯一id
	uid := hashid.HashID(id, hashid.ShareID)
	// 最终得到分享链接
	siteURL := model.GetSiteURL()
	sharePath, _ := url.Parse("/#/s/" + uid)
	shareURL := siteURL.ResolveReference(sharePath)

	return serializer.Response{
		Code: 0,
		Data: shareURL.String(),
	}
}
