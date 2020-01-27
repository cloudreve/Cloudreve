package model

import (
	"github.com/HFO4/cloudreve/pkg/hashid"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/jinzhu/gorm"
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
