package model

import (
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/jinzhu/gorm"
)

// Download 离线下载队列模型
type Download struct {
	gorm.Model
	Status   int    // 任务状态
	Type     int    // 任务类型
	Name     string // 任务文件名
	Size     uint64 // 文件大小
	PID      string // 任务ID
	Path     string `gorm:"type:text"` // 存储路径
	Attrs    string `gorm:"type:text"` // 任务状态属性
	FolderID uint   // 存储父目录ID
	UserID   uint   // 发起者UID
	TaskID   uint   // 对应的转存任务ID
}

// Create 创建离线下载记录
func (task *Download) Create() (uint, error) {
	if err := DB.Create(task).Error; err != nil {
		util.Log().Warning("无法插入离线下载记录, %s", err)
		return 0, err
	}
	return task.ID, nil
}
