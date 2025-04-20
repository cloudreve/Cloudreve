package model

import (
	"github.com/jinzhu/gorm"
)

// Tag 用户自定义标签
type Tag struct {
	gorm.Model
	Name       string // 标签名
	Icon       string // 图标标识
	Color      string // 图标颜色
	Type       int    // 标签类型（文件分类/目录直达）
	Expression string `gorm:"type:text"` // 搜索表表达式/直达路径
	UserID     uint   // 创建者ID
}

const (
	// FileTagType 文件分类标签
	FileTagType = iota
	// DirectoryLinkType 目录快捷方式标签
	DirectoryLinkType
)
