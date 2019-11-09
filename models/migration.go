package model

import (
	"cloudreve/pkg/conf"
	"cloudreve/pkg/util"
	"github.com/jinzhu/gorm"
	"github.com/mcuadros/go-version"
	"io/ioutil"
)

//执行数据迁移
func migration() {
	// 检查 version.lock 确认是否需要执行迁移
	// Debug 模式下一定会执行迁移
	if !conf.SystemConfig.Debug {
		if util.Exists("version.lock") {
			versionLock, _ := ioutil.ReadFile("version.lock")
			if version.Compare(string(versionLock), conf.BackendVersion, "=") {
				util.Log().Info("后端版本匹配，跳过数据库迁移")
				return
			}
		}
	}

	util.Log().Info("开始进行数据库自动迁移...")

	// 自动迁移模式
	DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&User{}, &Setting{})

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

	// 迁移完毕后写入版本锁 version.lock
	err = conf.WriteVersionLock()
	if err != nil {
		util.Log().Warning("无法写入版本控制锁 version.lock, ", err)
	}

}
