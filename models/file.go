package model

import (
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/jinzhu/gorm"
)

// File 文件
type File struct {
	// 表字段
	gorm.Model
	Name       string
	SourceName string
	UserID     uint `gorm:"index:user_id"`
	Size       uint64
	PicInfo    string
	FolderID   uint `gorm:"index:folder_id"`
	PolicyID   uint
	Dir        string `gorm:"size:65536"`
}

// Create 创建文件记录
func (file *File) Create() (uint, error) {
	if err := DB.Create(file).Error; err != nil {
		util.Log().Warning("无法插入文件记录, %s", err)
		return 0, err
	}
	return file.ID, nil
}

// GetFileByPathAndName 给定路径、文件名、用户ID，查找文件
func GetFileByPathAndName(path string, name string, uid uint) (File, error) {
	var file File
	result := DB.Where("user_id = ? AND dir = ? AND name=?", uid, path, name).First(&file)
	return file, result.Error
}

// GetChildFile 查找目录下子文件 TODO:test
func (folder *Folder) GetChildFile() ([]File, error) {
	var files []File
	result := DB.Where("folder_id = ?", folder.ID).Find(&files)
	return files, result.Error
}
