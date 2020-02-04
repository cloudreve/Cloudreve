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
	Options       string `json:"-",gorm:"type:text"`

	// 数据库忽略字段
	PolicyList        []uint      `gorm:"-"`
	OptionsSerialized GroupOption `gorm:"-"`
}

// GroupOption 用户组其他配置
type GroupOption struct {
	ArchiveDownload bool   `json:"archive_download,omitempty"` // 打包下载
	ArchiveTask     bool   `json:"archive_task,omitempty"`     // 在线压缩
	CompressSize    uint64 `json:"compress_size,omitempty"`    // 可压缩大小
	DecompressSize  uint64 `json:"decompress_size,omitempty"`
	OneTimeDownload bool   `json:"one_time_download,omitempty"`
	ShareDownload   bool   `json:"share_download,omitempty"`
	ShareFree       bool   `json:"share_free,omitempty"`
	Aria2           bool   `json:"aria2,omitempty"` // 离线下载
}

// GetAria2Option 获取用户离线下载设备
func (group *Group) GetAria2Option() [3]bool {
	if len(group.Aria2Option) != 5 {
		return [3]bool{false, false, false}
	}
	return [3]bool{
		group.Aria2Option[0] == '1',
		group.Aria2Option[2] == '1',
		group.Aria2Option[4] == '1',
	}
}

// GetGroupByID 用ID获取用户组
func GetGroupByID(ID interface{}) (Group, error) {
	var group Group
	result := DB.First(&group, ID)
	return group, result.Error
}

// AfterFind 找到用户组后的钩子，处理Policy列表
func (group *Group) AfterFind() (err error) {
	// 解析用户组策略列表
	if group.Policies != "" {
		err = json.Unmarshal([]byte(group.Policies), &group.PolicyList)
	}
	if err != nil {
		return err
	}

	// 解析用户组设置
	if group.Options != "" {
		err = json.Unmarshal([]byte(group.Options), &group.OptionsSerialized)
	}

	return err
}

// BeforeSave Save用户前的钩子
func (group *Group) BeforeSave() (err error) {
	err = group.SerializePolicyList()
	return err
}

//SerializePolicyList 将序列后的可选策略列表、配置写入数据库字段
// TODO 完善测试
func (group *Group) SerializePolicyList() (err error) {
	policies, err := json.Marshal(&group.PolicyList)
	group.Policies = string(policies)
	if err != nil {
		return err
	}

	optionsValue, err := json.Marshal(&group.OptionsSerialized)
	group.Options = string(optionsValue)
	return err
}
