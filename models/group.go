package model

import (
	"encoding/json"
	"github.com/jinzhu/gorm"
)

// Group 用户组模型
type Group struct {
	gorm.Model
	Name          string
	Policies      string
	MaxStorage    uint64
	ShareEnabled  bool
	WebDAVEnabled bool
	Aria2Option   string
	Color         string
	SpeedLimit    int

	// 数据库忽略字段
	PolicyList []uint `gorm:"-"`
}

// GetGroupByID 用ID获取用户组
func GetGroupByID(ID interface{}) (Group, error) {
	var group Group
	result := DB.First(&group, ID)
	return group, result.Error
}

// AfterFind 找到用户组后的钩子，处理Policy列表
func (group *Group) AfterFind() (err error) {
	// 解析用户设置到OptionsSerialized
	err = json.Unmarshal([]byte(group.Policies), &group.PolicyList)
	return err
}

// BeforeSave Save用户前的钩子
func (group *Group) BeforeSave() (err error) {
	err = group.SerializePolicyList()
	return err
}

//SerializePolicyList 将序列后的可选策略列表写入数据库字段
func (group *Group) SerializePolicyList() (err error) {
	optionsValue, err := json.Marshal(&group.PolicyList)
	group.Policies = string(optionsValue)
	return err
}
