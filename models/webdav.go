package model

import "github.com/jinzhu/gorm"

// Webdav 应用账户
type Webdav struct {
	gorm.Model
	Name     string // 应用名称
	Password string `gorm:"unique_index:password_only_on"` // 应用密码
	UserID   uint   `gorm:"unique_index:password_only_on"` // 用户ID
	Root     string `gorm:"type:text"`                     // 根目录
}

// GetWebdavByPassword 根据密码和用户查找Webdav应用
func GetWebdavByPassword(password string, uid uint) (*Webdav, error) {
	webdav := &Webdav{}
	res := DB.Where("user_id = ? and password = ?", uid, password).First(webdav)
	return webdav, res.Error
}
