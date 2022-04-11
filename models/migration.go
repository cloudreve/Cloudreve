package model

import (
	"context"
	"github.com/cloudreve/Cloudreve/v3/models/scripts/invoker"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/fatih/color"
	"github.com/hashicorp/go-version"
	"github.com/jinzhu/gorm"
	"sort"
	"strings"
)

// 是否需要迁移
func needMigration() bool {
	var setting Setting
	return DB.Where("name = ?", "db_version_"+conf.RequiredDBVersion).First(&setting).Error != nil
}

//执行数据迁移
func migration() {
	// 确认是否需要执行迁移
	if !needMigration() {
		util.Log().Info("数据库版本匹配，跳过数据库迁移")
		return

	}

	util.Log().Info("开始进行数据库初始化...")

	// 清除所有缓存
	if instance, ok := cache.Store.(*cache.RedisStore); ok {
		instance.DeleteAll()
	}

	// 自动迁移模式
	if conf.DatabaseConfig.Type == "mysql" {
		DB = DB.Set("gorm:table_options", "ENGINE=InnoDB")
	}

	DB.AutoMigrate(&User{}, &Setting{}, &Group{}, &Policy{}, &Folder{}, &File{}, &Share{},
		&Task{}, &Download{}, &Tag{}, &Webdav{}, &Node{})

	// 创建初始存储策略
	addDefaultPolicy()

	// 创建初始用户组
	addDefaultGroups()

	// 创建初始管理员账户
	addDefaultUser()

	// 创建初始节点
	addDefaultNode()

	// 向设置数据表添加初始设置
	addDefaultSettings()

	// 执行数据库升级脚本
	execUpgradeScripts()

	util.Log().Info("数据库初始化结束")

}

func addDefaultPolicy() {
	_, err := GetPolicyByID(uint(1))
	// 未找到初始存储策略时，则创建
	if gorm.IsRecordNotFoundError(err) {
		defaultPolicy := Policy{
			Name:               "默认存储策略",
			Type:               "local",
			MaxSize:            0,
			AutoRename:         true,
			DirNameRule:        "uploads/{uid}/{path}",
			FileNameRule:       "{uid}_{randomkey8}_{originname}",
			IsOriginLinkEnable: false,
		}
		if err := DB.Create(&defaultPolicy).Error; err != nil {
			util.Log().Panic("无法创建初始存储策略, %s", err)
		}
	}
}

func addDefaultSettings() {
	for _, value := range defaultSettings {
		DB.Where(Setting{Name: value.Name}).Create(&value)
	}
}

func addDefaultGroups() {
	_, err := GetGroupByID(1)
	// 未找到初始管理组时，则创建
	if gorm.IsRecordNotFoundError(err) {
		defaultAdminGroup := Group{
			Name:          "管理员",
			PolicyList:    []uint{1},
			MaxStorage:    1 * 1024 * 1024 * 1024,
			ShareEnabled:  true,
			WebDAVEnabled: true,
			OptionsSerialized: GroupOption{
				ArchiveDownload: true,
				ArchiveTask:     true,
				ShareDownload:   true,
				Aria2:           true,
			},
		}
		if err := DB.Create(&defaultAdminGroup).Error; err != nil {
			util.Log().Panic("无法创建管理用户组, %s", err)
		}
	}

	err = nil
	_, err = GetGroupByID(2)
	// 未找到初始注册会员时，则创建
	if gorm.IsRecordNotFoundError(err) {
		defaultAdminGroup := Group{
			Name:          "注册会员",
			PolicyList:    []uint{1},
			MaxStorage:    1 * 1024 * 1024 * 1024,
			ShareEnabled:  true,
			WebDAVEnabled: true,
			OptionsSerialized: GroupOption{
				ShareDownload: true,
			},
		}
		if err := DB.Create(&defaultAdminGroup).Error; err != nil {
			util.Log().Panic("无法创建初始注册会员用户组, %s", err)
		}
	}

	err = nil
	_, err = GetGroupByID(3)
	// 未找到初始游客用户组时，则创建
	if gorm.IsRecordNotFoundError(err) {
		defaultAdminGroup := Group{
			Name:       "游客",
			PolicyList: []uint{},
			Policies:   "[]",
			OptionsSerialized: GroupOption{
				ShareDownload: true,
			},
		}
		if err := DB.Create(&defaultAdminGroup).Error; err != nil {
			util.Log().Panic("无法创建初始游客用户组, %s", err)
		}
	}
}

func addDefaultUser() {
	_, err := GetUserByID(1)
	password := util.RandStringRunes(8)

	// 未找到初始用户时，则创建
	if gorm.IsRecordNotFoundError(err) {
		defaultUser := NewUser()
		defaultUser.Email = "admin@cloudreve.org"
		defaultUser.Nick = "admin"
		defaultUser.Status = Active
		defaultUser.GroupID = 1
		err := defaultUser.SetPassword(password)
		if err != nil {
			util.Log().Panic("无法创建密码, %s", err)
		}
		if err := DB.Create(&defaultUser).Error; err != nil {
			util.Log().Panic("无法创建初始用户, %s", err)
		}

		c := color.New(color.FgWhite).Add(color.BgBlack).Add(color.Bold)
		util.Log().Info("初始管理员账号：" + c.Sprint("admin@cloudreve.org"))
		util.Log().Info("初始管理员密码：" + c.Sprint(password))
	}
}

func addDefaultNode() {
	_, err := GetNodeByID(1)

	if gorm.IsRecordNotFoundError(err) {
		defaultAdminGroup := Node{
			Name:   "主机（本机）",
			Status: NodeActive,
			Type:   MasterNodeType,
			Aria2OptionsSerialized: Aria2Option{
				Interval: 10,
				Timeout:  10,
			},
		}
		if err := DB.Create(&defaultAdminGroup).Error; err != nil {
			util.Log().Panic("无法创建初始节点记录, %s", err)
		}
	}
}

func execUpgradeScripts() {
	s := invoker.ListPrefix("UpgradeTo")
	versions := make([]*version.Version, len(s))
	for i, raw := range s {
		v, _ := version.NewVersion(strings.TrimPrefix(raw, "UpgradeTo"))
		versions[i] = v
	}
	sort.Sort(version.Collection(versions))

	for i := 0; i < len(versions); i++ {
		invoker.RunDBScript("UpgradeTo"+versions[i].String(), context.Background())
	}
}
