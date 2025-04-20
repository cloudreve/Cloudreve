package model

import (
	"github.com/jinzhu/gorm"
)

// Task 任务模型
type Task struct {
	gorm.Model
	Status   int    // 任务状态
	Type     int    // 任务类型
	UserID   uint   // 发起者UID，0表示为系统发起
	Progress int    // 进度
	Error    string `gorm:"type:text"` // 错误信息
	Props    string `gorm:"type:text"` // 任务属性
}
