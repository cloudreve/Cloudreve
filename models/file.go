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

// Create 创建文件记录
func (file *File) Create() (uint, error) {
	if err := DB.Create(file).Error; err != nil {
		return 0, err
	}
	return file.ID, nil
}

// GetFileByPathAndName 给定路径、文件名、用户ID，查找文件
func GetFileByPathAndName(path string, name string, uid uint) (File, error) {
	var file File
	result := DB.Where("user_id = ? AND dir = ? AND name=?", uid, path, name).Find(&file)
	return file, result.Error
}
