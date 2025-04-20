package model

import (
	"time"

	"github.com/jinzhu/gorm"
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
	PreviewEnabled  bool       // 是否允许直接预览
	SourceName      string     `gorm:"index:source"` // 用于搜索的字段

	// 数据库忽略字段
	User   User   `gorm:"PRELOAD:false,association_autoupdate:false"`
	File   File   `gorm:"PRELOAD:false,association_autoupdate:false"`
	Folder Folder `gorm:"PRELOAD:false,association_autoupdate:false"`
}
