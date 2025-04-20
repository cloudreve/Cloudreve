package model

import (
	"github.com/jinzhu/gorm"
)

const (
	// Active 账户正常状态
	Active = iota
	// NotActivicated 未激活
	NotActivicated
	// Baned 被封禁
	Baned
	// OveruseBaned 超额使用被封禁
	OveruseBaned
)

// User 用户模型
type User struct {
	// 表字段
	gorm.Model
	Email     string `gorm:"type:varchar(100);unique_index"`
	Nick      string `gorm:"size:50"`
	Password  string `json:"-"`
	Status    int
	GroupID   uint
	Storage   uint64
	TwoFactor string
	Avatar    string
	Options   string `json:"-" gorm:"size:4294967295"`
	Authn     string `gorm:"size:4294967295"`

	// 关联模型
	Group  Group  `gorm:"save_associations:false:false"`
	Policy Policy `gorm:"PRELOAD:false,association_autoupdate:false"`

	// 数据库忽略字段
	OptionsSerialized UserOption `gorm:"-"`
}

// UserOption 用户个性化配置字段
type UserOption struct {
	ProfileOff     bool   `json:"profile_off,omitempty"`
	PreferredTheme string `json:"preferred_theme,omitempty"`
}
