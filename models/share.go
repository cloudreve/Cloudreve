package model

import (
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
