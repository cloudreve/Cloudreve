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
	SpeedLimit    int
	Options       string `json:"-" gorm:"size:4294967295"`

	// 数据库忽略字段
	PolicyList        []uint      `gorm:"-"`
	OptionsSerialized GroupOption `gorm:"-"`
	GroupFolder       *Folder     `gorm:"-"`
}

// GroupOption 用户组其他配置
type GroupOption struct {
	ArchiveDownload    bool                   `json:"archive_download,omitempty"` // 打包下载
	ArchiveTask        bool                   `json:"archive_task,omitempty"`     // 在线压缩
	CompressSize       uint64                 `json:"compress_size,omitempty"`    // 可压缩大小
	DecompressSize     uint64                 `json:"decompress_size,omitempty"`
	OneTimeDownload    bool                   `json:"one_time_download,omitempty"`
	ShareDownload      bool                   `json:"share_download,omitempty"`
	Aria2              bool                   `json:"aria2,omitempty"`         // 离线下载
	Aria2Options       map[string]interface{} `json:"aria2_options,omitempty"` // 离线下载用户组配置
	SourceBatchSize    int                    `json:"source_batch,omitempty"`
	RedirectedSource   bool                   `json:"redirected_source,omitempty"`
	Aria2BatchSize     int                    `json:"aria2_batch,omitempty"`
	AdvanceDelete      bool                   `json:"advance_delete,omitempty"`
	WebDAVProxy        bool                   `json:"webdav_proxy,omitempty"`
	GroupFolderEnabled bool                   `json:"group_folder_enabled,omitempty"`
	GroupFolder        string                 `json:"group_folder,omitempty"`
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

func (group *Group) AfterSave() (err error) {
	err = json.Unmarshal([]byte(group.Options), &group.OptionsSerialized)
	if err != nil {
		return err
	}

	if group.OptionsSerialized.GroupFolderEnabled {
		err = group.createGroupFolder(group.OptionsSerialized.GroupFolder)
	} else {
		err = nil
	}

	return err
}

// SerializePolicyList 将序列后的可选策略列表、配置写入数据库字段
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

func (group *Group) getGroupFolder() (*Folder, bool) {
	var folder Folder
	result := DB.Where("owner_id = ?", -int(group.ID)).First(&folder)
	return &folder, !result.RecordNotFound()
}

// createGroupFolder 创建用户组文件夹，如果已存在则更新
func (group *Group) createGroupFolder(name string) error {
	tx := DB.Begin()
	folder, ok := group.getGroupFolder()
	if !ok {
		folder = &Folder{
			OwnerID: -int(group.ID),
			Name:    name,
		}
		if err := tx.Create(folder).Error; err != nil {
			return err
		}
	} else {
		folder.Name = name
		if err := tx.Save(folder).Error; err != nil {
			return err
		}
	}

	group.GroupFolder = folder
	return tx.Commit().Error
}

// deleteGroupFolder 删除用户组文件夹
func (group *Group) deleteGroupFolder() {
	tx := DB.Begin()
	var folder Folder
	if err := tx.Where("owner_id = ?", -int(group.ID)).First(&folder).Error; err != nil {
		return
	}

	if err := tx.Delete(&folder).Error; err != nil {
		return
	}

	tx.Commit()
}
