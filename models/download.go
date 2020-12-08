package model

import (
	"encoding/json"

	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/rpc"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/jinzhu/gorm"
)

// Download 离线下载队列模型
type Download struct {
	gorm.Model
	Status         int    // 任务状态
	Type           int    // 任务类型
	Source         string `gorm:"type:text"` // 文件下载地址
	TotalSize      uint64 // 文件大小
	DownloadedSize uint64 // 文件大小
	GID            string `gorm:"size:32,index:gid"` // 任务ID
	Speed          int    // 下载速度
	Parent         string `gorm:"type:text"`  // 存储目录
	Attrs          string `gorm:"size:65535"` // 任务状态属性
	Error          string `gorm:"type:text"`  // 错误描述
	Dst            string `gorm:"type:text"`  // 用户文件系统存储父目录路径
	UserID         uint   // 发起者UID
	TaskID         uint   // 对应的转存任务ID

	// 关联模型
	User *User `gorm:"PRELOAD:false,association_autoupdate:false"`

	// 数据库忽略字段
	StatusInfo rpc.StatusInfo `gorm:"-"`
	Task       *Task          `gorm:"-"`
}

// AfterFind 找到下载任务后的钩子，处理Status结构
func (task *Download) AfterFind() (err error) {
	// 解析状态
	if task.Attrs != "" {
		err = json.Unmarshal([]byte(task.Attrs), &task.StatusInfo)
	}

	if task.TaskID != 0 {
		task.Task, _ = GetTasksByID(task.TaskID)
	}

	return err
}

// BeforeSave Save下载任务前的钩子
func (task *Download) BeforeSave() (err error) {
	// 解析状态
	if task.Attrs != "" {
		err = json.Unmarshal([]byte(task.Attrs), &task.StatusInfo)
	}
	return err
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

// GetDownloadsByStatusAndUser 根据状态检索和用户ID下载
// page 为 0 表示列出所有，非零时分页
func GetDownloadsByStatusAndUser(page, uid uint, status ...int) []Download {
	var tasks []Download
	dbChain := DB
	if page > 0 {
		dbChain = dbChain.Limit(10).Offset((page - 1) * 10).Order("updated_at DESC")
	}
	dbChain.Where("user_id = ? and status in (?)", uid, status).Find(&tasks)
	return tasks
}

// GetDownloadByGid 根据GID和用户ID查找下载
func GetDownloadByGid(gid string, uid uint) (*Download, error) {
	download := &Download{}
	result := DB.Where("user_id = ? and g_id = ?", uid, gid).First(download)
	return download, result.Error
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

// Delete 删除离线下载记录
func (download *Download) Delete() error {
	return DB.Model(download).Delete(download).Error
}
