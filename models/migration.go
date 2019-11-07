package model

import (
	"Cloudreve/pkg/util"
	"github.com/jinzhu/gorm"
)

//执行数据迁移
func migration() {
	// 自动迁移模式
	DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&User{})

	// 添加初始用户
	_, err := GetUser(1)
	if gorm.IsRecordNotFoundError(err) {
		defaultUser := NewUser()
		//TODO 动态生成密码
		defaultUser.Email = "admin@cloudreve.org"
		defaultUser.Nick = "admin"
		defaultUser.Status = Active
		defaultUser.Group = 1
		defaultUser.PrimaryGroup = 1
		err := defaultUser.SetPassword("admin")
		if err != nil {
			util.Log().Panic("无法创建密码, ", err)
		}
		if err := DB.Create(&defaultUser).Error; err != nil {
			util.Log().Panic("无法创建初始用户, ", err)
		}
	}

}
