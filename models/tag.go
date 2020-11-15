package model

import (
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
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

// Create 创建标签记录
func (tag *Tag) Create() (uint, error) {
	if err := DB.Create(tag).Error; err != nil {
		util.Log().Warning("无法插入离线下载记录, %s", err)
		return 0, err
	}
	return tag.ID, nil
}

// DeleteTagByID 根据给定ID和用户ID删除标签
func DeleteTagByID(id, uid uint) error {
	result := DB.Where("id = ? and user_id = ?", id, uid).Delete(&Tag{})
	return result.Error
}

// GetTagsByUID 根据用户ID查找标签
func GetTagsByUID(uid uint) ([]Tag, error) {
	var tag []Tag
	result := DB.Where("user_id = ?", uid).Find(&tag)
	return tag, result.Error
}

// GetTagsByID 根据ID查找标签
func GetTagsByID(id, uid uint) (*Tag, error) {
	var tag Tag
	result := DB.Where("user_id = ? and id = ?", uid, id).First(&tag)
	return &tag, result.Error
}
