package model

import "github.com/jinzhu/gorm"

// Folder 目录
type Folder struct {
	// 表字段
	gorm.Model
	Name             string
	ParentID         uint
	Position         string `gorm:"size:65536"`
	OwnerID          uint
	PositionAbsolute string `gorm:"size:65536"`

	// 关联模型
	OptionsSerialized PolicyOption `gorm:"-"`
}
