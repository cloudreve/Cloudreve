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

// settingCache 设置项缓存
var settingCache = make(map[string]string)

// GetSettingByName 用 Name 获取设置值
func GetSettingByName(name string) (string, error) {
	var setting Setting

	// 优先从缓存中查找
	if optionValue, ok := settingCache[name]; ok {
		return optionValue, nil
	} else {
		// 尝试数据库中查找
		result := DB.Where("name = ?", name).First(&setting)
		if result.Error == nil {
			settingCache[setting.Name] = setting.Value
			return setting.Value, nil
		}
		return "", result.Error
	}

}
