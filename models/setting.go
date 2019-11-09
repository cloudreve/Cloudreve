package model

import (
	"github.com/jinzhu/gorm"
)

// Setting 系统设置模型
type Setting struct {
	gorm.Model
	Type  string `gorm:"not null"`
	Name  string `gorm:"unique;not null;index:setting_key"`
	Value string `gorm:"size:‎65535"`
}

// GetSettingByName 用 Name 获取设置值
func GetSettingByName(name string) (Setting, error) {
	var setting Setting

	// 优先尝试数据库中查找
	result := DB.Where("name = ?", name).First(&setting)
	if result.Error == nil {
		return setting, nil
	}

	return setting, result.Error
}
