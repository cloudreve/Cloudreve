package model

import (
	"github.com/jinzhu/gorm"
)

// Report 举报模型
type Report struct {
	gorm.Model
	ShareID     uint   `gorm:"index:share_id"` // 对应分享ID
	Reason      int    // 举报原因
	Description string // 补充描述

	// 关联模型
	Share Share `gorm:"save_associations:false:false"`
}

// Create 创建举报
func (report *Report) Create() error {
	return DB.Create(report).Error
}
