package model

import (
	"github.com/jinzhu/gorm"
)

// Webdav 应用账户
type Webdav struct {
	gorm.Model
	Name     string // 应用名称
	Password string `gorm:"unique_index:password_only_on"` // 应用密码
	UserID   uint   `gorm:"unique_index:password_only_on"` // 用户ID
	Root     string `gorm:"type:text"`                     // 根目录
	Readonly bool   `gorm:"type:bool"`                     // 是否只读
	UseProxy bool   `gorm:"type:bool"`                     // 是否进行反代
}

// Create 创建账户
func (webdav *Webdav) Create() (uint, error) {
	if err := DB.Create(webdav).Error; err != nil {
		return 0, err
	}
	return webdav.ID, nil
}

// GetWebdavByPassword 根据密码和用户查找Webdav应用
func GetWebdavByPassword(password string, uid uint) (*Webdav, error) {
	webdav := &Webdav{}
	res := DB.Where("user_id = ? and password = ?", uid, password).First(webdav)
	return webdav, res.Error
}

// ListWebDAVAccounts 列出用户的所有账号
func ListWebDAVAccounts(uid uint) []Webdav {
	var accounts []Webdav
	DB.Where("user_id = ?", uid).Order("created_at desc").Find(&accounts)
	return accounts
}

// DeleteWebDAVAccountByID 根据账户ID和UID删除账户
func DeleteWebDAVAccountByID(id, uid uint) {
	DB.Where("user_id = ? and id = ?", uid, id).Delete(&Webdav{})
}

// UpdateWebDAVAccountByID 根据账户ID和UID更新账户
func UpdateWebDAVAccountByID(id, uid uint, updates map[string]interface{}) {
	DB.Model(&Webdav{Model: gorm.Model{ID: id}, UserID: uid}).Updates(updates)
}

// UpdateWebDAVAccountReadonlyByID 根据账户ID和UID更新账户的只读性
func UpdateWebDAVAccountReadonlyByID(id, uid uint, readonly bool) {
	DB.Model(&Webdav{Model: gorm.Model{ID: id}, UserID: uid}).UpdateColumn("readonly", readonly)
}
