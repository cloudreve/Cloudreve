package model

import (
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/jinzhu/gorm"
)

// Folder 目录
type Folder struct {
	// 表字段
	gorm.Model
	Name     string
	ParentID uint   `gorm:"index:parent_id"`
	Position string `gorm:"size:65536"`
	OwnerID  uint   `gorm:"index:owner_id"`
	//PositionAbsolute string `gorm:"size:65536"`
}

// Create 创建目录
func (folder *Folder) Create() (uint, error) {
	if err := DB.Create(folder).Error; err != nil {
		util.Log().Warning("无法插入目录记录, %s", err)
		return 0, err
	}
	return folder.ID, nil
}

// GetFolderByPath 根据绝对路径和UID查找目录
func GetFolderByPath(path string, uid uint) (Folder, error) {
	var folder Folder
	result := DB.Where("owner_id = ? AND position_absolute = ?", uid, path).First(&folder)
	return folder, result.Error
}

// GetChildFolder 查找子目录
func (folder *Folder) GetChildFolder() ([]Folder, error) {
	var folders []Folder
	result := DB.Where("parent_id = ?", folder.ID).Find(&folders)
	return folders, result.Error
}

// GetRecursiveChildFolder 查找所有递归子目录
func GetRecursiveChildFolder(dirs []string, uid uint) ([]Folder, error) {
	folders := make([]Folder, 0, len(dirs))
	search := util.BuildRegexp(dirs, "^", "/", "|")
	result := DB.Where("(owner_id = ? and position_absolute REGEXP ?) or position_absolute in (?)", uid, search, dirs).Find(&folders)
	return folders, result.Error
}

// DeleteFolderByIDs 根据给定ID批量删除目录记录
func DeleteFolderByIDs(ids []uint) error {
	result := DB.Where("id in (?)", ids).Delete(&Folder{})
	return result.Error
}
