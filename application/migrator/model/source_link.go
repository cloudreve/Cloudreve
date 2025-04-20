package model

import (
	"github.com/jinzhu/gorm"
)

// SourceLink represent a shared file source link
type SourceLink struct {
	gorm.Model
	FileID    uint   // corresponding file ID
	Name      string // name of the file while creating the source link, for annotation
	Downloads int    // 下载数

	// 关联模型
	File File `gorm:"save_associations:false:false"`
}
