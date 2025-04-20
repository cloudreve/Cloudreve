package model

import (
	"github.com/jinzhu/gorm"
)

// Folder 目录
type Folder struct {
	// 表字段
	gorm.Model
	Name     string `gorm:"unique_index:idx_only_one_name"`
	ParentID *uint  `gorm:"index:parent_id;unique_index:idx_only_one_name"`
	OwnerID  uint   `gorm:"index:owner_id"`

	// 数据库忽略字段
	Position      string `gorm:"-"`
	WebdavDstName string `gorm:"-"`
}
