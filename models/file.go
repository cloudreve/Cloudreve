package model

import "github.com/jinzhu/gorm"

// File 文件
type File struct {
	// 表字段
	gorm.Model
	Name       string
	SourceName string
	UserID     uint
	Size       uint64
	PicInfo    string
	FolderID   uint
	PolicyID   uint
	Dir        string `gorm:"size:65536"`
}
