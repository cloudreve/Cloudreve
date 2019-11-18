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

// GetFolderByPath 根据绝对路径和UID查找目录
func GetFolderByPath(path string, uid uint) (Folder, error) {
	var folder Folder
	result := DB.Where("owner = ? AND position_absolute = ?", uid, path).Find(&folder)
	return folder, result.Error
}
