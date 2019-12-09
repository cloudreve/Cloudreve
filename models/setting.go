package model

import (
	"github.com/HFO4/cloudreve/pkg/cache"
	"github.com/jinzhu/gorm"
)

// Setting 系统设置模型
type Setting struct {
	gorm.Model
	Type  string `gorm:"not null"`
	Name  string `gorm:"unique;not null;index:setting_key"`
	Value string `gorm:"size:‎65535"`
}

// IsTrueVal 返回设置的值是否为真
func IsTrueVal(val string) bool {
	return val == "1" || val == "true"
}

// GetSettingByName 用 Name 获取设置值
func GetSettingByName(name string) string {
	var setting Setting

	// 优先从缓存中查找
	cacheKey := "setting_" + name
	if optionValue, ok := cache.Get(cacheKey); ok {
		return optionValue.(string)
	}
	// 尝试数据库中查找
	result := DB.Where("name = ?", name).First(&setting)
	if result.Error == nil {
		_ = cache.Set(cacheKey, setting.Value)
		return setting.Value
	}
	return ""
}

// GetSettingByNames 用多个 Name 获取设置值
// TODO 其他设置获取也使用缓存
func GetSettingByNames(names []string) map[string]string {
	var queryRes []Setting
	res, miss := cache.GetsSettingByName(names)

	DB.Where("name IN (?)", miss).Find(&queryRes)
	for _, setting := range queryRes {
		res[setting.Name] = setting.Value
	}

	_ = cache.SetSettings(res)
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
