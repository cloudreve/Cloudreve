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

// ClearCache 清空设置缓存
func ClearCache() {
	settingCache = make(map[string]string)
}

// IsTrueVal 返回设置的值是否为真
func IsTrueVal(val string) bool {
	return val == "1" || val == "true"
}

// GetSettingByName 用 Name 获取设置值
func GetSettingByName(name string) string {
	var setting Setting

	// 优先从缓存中查找
	if optionValue, ok := settingCache[name]; ok {
		return optionValue
	}
	// 尝试数据库中查找
	result := DB.Where("name = ?", name).First(&setting)
	if result.Error == nil {
		settingCache[setting.Name] = setting.Value
		return setting.Value
	}
	return ""
}

// GetSettingByNames 用多个 Name 获取设置值
func GetSettingByNames(names []string) map[string]string {
	var queryRes []Setting
	res := make(map[string]string)

	DB.Where("name IN (?)", names).Find(&queryRes)
	for _, setting := range queryRes {
		res[setting.Name] = setting.Value
	}

	return res
}

// GetSettingByType 获取一个或多个分组的所有设置值
func GetSettingByType(types []string) map[string]string {
	var queryRes []Setting
	res := make(map[string]string)

	DB.Where("type IN (?)", types).Find(&queryRes)
	for _, setting := range queryRes {
		res[setting.Name] = setting.Value
	}

	return res
}
