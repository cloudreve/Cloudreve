package model

import (
	"errors"
	"fmt"
	"github.com/HFO4/cloudreve/pkg/cache"
	"github.com/HFO4/cloudreve/pkg/hashid"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"math"
	"time"
)

// Share 分享模型
type Share struct {
	gorm.Model
	Password        string     // 分享密码，空值为非加密分享
	IsDir           bool       // 原始资源是否为目录
	UserID          uint       // 创建用户ID
	SourceID        uint       // 原始资源ID
	Views           int        // 浏览数
	Downloads       int        // 下载数
	RemainDownloads int        // 剩余下载配额，负值标识无限制
	Expires         *time.Time // 过期时间，空值表示无过期时间
	Score           int        // 每人次下载扣除积分
	PreviewEnabled  bool       // 是否允许直接预览

	// 数据库忽略字段
	User   User   `gorm:"PRELOAD:false,association_autoupdate:false"`
	File   File   `gorm:"PRELOAD:false,association_autoupdate:false"`
	Folder Folder `gorm:"PRELOAD:false,association_autoupdate:false"`
}

// Create 创建分享
// TODO 测试
func (share *Share) Create() (uint, error) {
	if err := DB.Create(share).Error; err != nil {
		util.Log().Warning("无法插入数据库记录, %s", err)
		return 0, err
	}
	return share.ID, nil
}

// GetShareByHashID 根据HashID查找分享
// TODO 测试
func GetShareByHashID(hashID string) *Share {
	id, err := hashid.DecodeHashID(hashID, hashid.ShareID)
	if err != nil {
		return nil
	}
	var share Share
	result := DB.First(&share, id)
	if result.Error != nil {
		return nil
	}

	return &share
}

// IsAvailable 返回此分享是否可用（是否过期）
// TODO 测试
func (share *Share) IsAvailable() bool {
	if share.RemainDownloads == 0 {
		return false
	}
	if share.Expires != nil && time.Now().After(*share.Expires) {
		return false
	}

	// 检查源对象是否存在
	var sourceID uint
	if share.IsDir {
		folder := share.GetSourceFolder()
		sourceID = folder.ID
	} else {
		file := share.GetSourceFile()
		sourceID = file.ID
	}
	if sourceID == 0 {
		return false
	}

	return true
}

// GetCreator 获取分享的创建者
func (share *Share) GetCreator() *User {
	if share.User.ID == 0 {
		share.User, _ = GetUserByID(share.UserID)
	}
	return &share.User
}

// GetSource 返回源对象
func (share *Share) GetSource() interface{} {
	if share.IsDir {
		return share.GetSourceFolder()
	}
	return share.GetSourceFile()
}

// GetSourceFolder 获取源目录
func (share *Share) GetSourceFolder() *Folder {
	if share.Folder.ID == 0 {
		folders, _ := GetFoldersByIDs([]uint{share.SourceID}, share.UserID)
		if len(folders) > 0 {
			share.Folder = folders[0]
		}
	}
	return &share.Folder
}

// GetSourceFile 获取源文件
func (share *Share) GetSourceFile() *File {
	if share.File.ID == 0 {
		files, _ := GetFilesByIDs([]uint{share.SourceID}, share.UserID)
		if len(files) > 0 {
			share.File = files[0]
		}
	}
	return &share.File
}

// CanBeDownloadBy 返回此分享是否可以被给定用户下载
func (share *Share) CanBeDownloadBy(user *User) error {
	// 用户组权限
	if !user.Group.OptionsSerialized.ShareDownloadEnabled {
		if user.IsAnonymous() {
			return errors.New("未登录用户无法下载")
		}
		return errors.New("您当前的用户组无权下载")
	}

	// 需要积分但未登录
	if share.Score > 0 && user.IsAnonymous() {
		return errors.New("未登录用户无法下载")
	}

	return nil
}

// WasDownloadedBy 返回分享是否已被用户下载过
func (share *Share) WasDownloadedBy(user *User, c *gin.Context) (exist bool) {
	if user.IsAnonymous() {
		exist = util.GetSession(c, fmt.Sprintf("share_%d_%d", share.ID, user.ID)) != nil
	} else {
		_, exist = cache.Get(fmt.Sprintf("share_%d_%d", share.ID, user.ID))
	}

	return exist
}

// DownloadBy 增加下载次数、检查积分等，匿名用户不会缓存
func (share *Share) DownloadBy(user *User, c *gin.Context) error {
	if !share.WasDownloadedBy(user, c) {
		if err := share.Purchase(user); err != nil {
			return err
		}
		share.Downloaded()
		if !user.IsAnonymous() {
			cache.Set(fmt.Sprintf("share_%d_%d", share.ID, user.ID), true,
				GetIntSetting("share_download_session_timeout", 2073600))
		} else {
			util.SetSession(c, map[string]interface{}{fmt.Sprintf("share_%d_%d", share.ID, user.ID): true})
		}
	}
	return nil
}

// Purchase 使用积分购买分享
func (share *Share) Purchase(user *User) error {
	// 不需要付积分
	if share.Score == 0 || user.Group.OptionsSerialized.ShareFreeEnabled || user.ID == share.UserID {
		return nil
	}

	ok := user.PayScore(share.Score)
	if !ok {
		return errors.New("积分不足")
	}

	scoreRate := GetIntSetting("share_score_rate", 100)
	gainedScore := int(math.Ceil(float64(share.Score*scoreRate) / 100))
	share.GetCreator().AddScore(gainedScore)

	return nil
}

// Viewed 增加访问次数
func (share *Share) Viewed() {
	share.Views++
	DB.Model(share).UpdateColumn("views", gorm.Expr("views + ?", 1))
}

// Downloaded 增加下载次数
func (share *Share) Downloaded() {
	share.Downloads++
	if share.RemainDownloads > 0 {
		share.RemainDownloads--
	}
	DB.Model(share).Updates(map[string]interface{}{
		"downloads":        share.Downloads,
		"remain_downloads": share.RemainDownloads,
	})
}
