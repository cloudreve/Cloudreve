package model

import (
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/jinzhu/gorm"
)

// Download 离线下载队列模型
type Download struct {
	gorm.Model
	Status         int    // 任务状态
	Type           int    // 任务类型
	Source         string // 文件下载地址
	TotalSize      uint64 // 文件大小
	DownloadedSize uint64 // 文件大小
	GID            string // 任务ID
	Speed          int    // 下载速度
	Path           string `gorm:"type:text"` // 存储路径
	Parent         string `gorm:"type:text"` // 存储目录
	Attrs          string `gorm:"type:text"` // 任务状态属性
	Error          string `gorm:"type:text"` // 错误描述
	Dst            string `gorm:"type:text"` // 用户文件系统存储父目录路径
	UserID         uint   // 发起者UID
	TaskID         uint   // 对应的转存任务ID

	// 关联模型
	User *User `gorm:"PRELOAD:false,association_autoupdate:false"`
}

// Create 创建离线下载记录
func (task *Download) Create() (uint, error) {
	if err := DB.Create(task).Error; err != nil {
		util.Log().Warning("无法插入离线下载记录, %s", err)
		return 0, err
	}
	return task.ID, nil
}

// Save 更新
func (task *Download) Save() error {
	if err := DB.Save(task).Error; err != nil {
		util.Log().Warning("无法更新离线下载记录, %s", err)
		return err
	}
	return nil
}

// GetDownloadsByStatus 根据状态检索下载
func GetDownloadsByStatus(status ...int) []Download {
	var tasks []Download
	DB.Where("status in (?)", status).Find(&tasks)
	return tasks
}

// GetOwner 获取下载任务所属用户
func (task *Download) GetOwner() *User {
	if task.User == nil {
		if user, err := GetUserByID(task.UserID); err == nil {
			return &user
		}
	}
	return task.User
}
