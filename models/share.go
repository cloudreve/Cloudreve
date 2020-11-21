package model

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// Share 分享模型
type Share struct {
	gorm.Model
	Password        string     // 分享密码，空值为非加密分享
	IsDir           bool       // 原始资源是否为目录
	UserID          uint       // 创建用户ID
	SourceID        uint       // 原始资源ID
	Views           int        // 浏览数
	Downloads       int        // 下载数
	RemainDownloads int        // 剩余下载配额，负值标识无限制
	Expires         *time.Time // 过期时间，空值表示无过期时间
	PreviewEnabled  bool       // 是否允许直接预览
	SourceName      string     `gorm:"index:source"` // 用于搜索的字段

	// 数据库忽略字段
	User   User   `gorm:"PRELOAD:false,association_autoupdate:false"`
	File   File   `gorm:"PRELOAD:false,association_autoupdate:false"`
	Folder Folder `gorm:"PRELOAD:false,association_autoupdate:false"`
}

// Create 创建分享
func (share *Share) Create() (uint, error) {
	if err := DB.Create(share).Error; err != nil {
		util.Log().Warning("无法插入数据库记录, %s", err)
		return 0, err
	}
	return share.ID, nil
}

// GetShareByHashID 根据HashID查找分享
func GetShareByHashID(hashID string) *Share {
	id, err := hashid.DecodeHashID(hashID, hashid.ShareID)
	if err != nil {
		return nil
	}
	var share Share
	result := DB.First(&share, id)
	if result.Error != nil {
		return nil
	}

	return &share
}

// IsAvailable 返回此分享是否可用（是否过期）
func (share *Share) IsAvailable() bool {
	if share.RemainDownloads == 0 {
		return false
	}
	if share.Expires != nil && time.Now().After(*share.Expires) {
		return false
	}

	// 检查创建者状态
	if share.Creator().Status != Active {
		return false
	}

	// 检查源对象是否存在
	var sourceID uint
	if share.IsDir {
		folder := share.SourceFolder()
		sourceID = folder.ID
	} else {
		file := share.SourceFile()
		sourceID = file.ID
	}
	if sourceID == 0 {
		// TODO 是否要在这里删除这个无效分享？
		return false
	}

	return true
}

// Creator 获取分享的创建者
func (share *Share) Creator() *User {
	if share.User.ID == 0 {
		share.User, _ = GetUserByID(share.UserID)
	}
	return &share.User
}

// Source 返回源对象
func (share *Share) Source() interface{} {
	if share.IsDir {
		return share.SourceFolder()
	}
	return share.SourceFile()
}

// SourceFolder 获取源目录
func (share *Share) SourceFolder() *Folder {
	if share.Folder.ID == 0 {
		folders, _ := GetFoldersByIDs([]uint{share.SourceID}, share.UserID)
		if len(folders) > 0 {
			share.Folder = folders[0]
		}
	}
	return &share.Folder
}

// SourceFile 获取源文件
func (share *Share) SourceFile() *File {
	if share.File.ID == 0 {
		files, _ := GetFilesByIDs([]uint{share.SourceID}, share.UserID)
		if len(files) > 0 {
			share.File = files[0]
		}
	}
	return &share.File
}

// CanBeDownloadBy 返回此分享是否可以被给定用户下载
func (share *Share) CanBeDownloadBy(user *User) error {
	// 用户组权限
	if !user.Group.OptionsSerialized.ShareDownload {
		if user.IsAnonymous() {
			return errors.New("未登录用户无法下载")
		}
		return errors.New("您当前的用户组无权下载")
	}
	return nil
}

// WasDownloadedBy 返回分享是否已被用户下载过
func (share *Share) WasDownloadedBy(user *User, c *gin.Context) (exist bool) {
	if user.IsAnonymous() {
		exist = util.GetSession(c, fmt.Sprintf("share_%d_%d", share.ID, user.ID)) != nil
	} else {
		_, exist = cache.Get(fmt.Sprintf("share_%d_%d", share.ID, user.ID))
	}

	return exist
}

// DownloadBy 增加下载次数，匿名用户不会缓存
func (share *Share) DownloadBy(user *User, c *gin.Context) error {
	if !share.WasDownloadedBy(user, c) {
		share.Downloaded()
		if !user.IsAnonymous() {
			cache.Set(fmt.Sprintf("share_%d_%d", share.ID, user.ID), true,
				GetIntSetting("share_download_session_timeout", 2073600))
		} else {
			util.SetSession(c, map[string]interface{}{fmt.Sprintf("share_%d_%d", share.ID, user.ID): true})
		}
	}
	return nil
}

// Viewed 增加访问次数
func (share *Share) Viewed() {
	share.Views++
	DB.Model(share).UpdateColumn("views", gorm.Expr("views + ?", 1))
}

// Downloaded 增加下载次数
func (share *Share) Downloaded() {
	share.Downloads++
	if share.RemainDownloads > 0 {
		share.RemainDownloads--
	}
	DB.Model(share).Updates(map[string]interface{}{
		"downloads":        share.Downloads,
		"remain_downloads": share.RemainDownloads,
	})
}

// Update 更新分享属性
func (share *Share) Update(props map[string]interface{}) error {
	return DB.Model(share).Updates(props).Error
}

// Delete 删除分享
func (share *Share) Delete() error {
	return DB.Model(share).Delete(share).Error
}

// DeleteShareBySourceIDs 根据原始资源类型和ID删除文件
func DeleteShareBySourceIDs(sources []uint, isDir bool) error {
	return DB.Where("source_id in (?) and is_dir = ?", sources, isDir).Delete(&Share{}).Error
}

// ListShares 列出UID下的分享
func ListShares(uid uint, page, pageSize int, order string, publicOnly bool) ([]Share, int) {
	var (
		shares []Share
		total  int
	)
	dbChain := DB
	dbChain = dbChain.Where("user_id = ?", uid)
	if publicOnly {
		dbChain = dbChain.Where("password = ?", "")
	}

	// 计算总数用于分页
	dbChain.Model(&Share{}).Count(&total)

	// 查询记录
	dbChain.Limit(pageSize).Offset((page - 1) * pageSize).Order(order).Find(&shares)
	return shares, total
}

// SearchShares 根据关键字搜索分享
func SearchShares(page, pageSize int, order, keywords string) ([]Share, int) {
	var (
		shares []Share
		total  int
	)

	keywordList := strings.Split(keywords, " ")
	availableList := make([]string, 0, len(keywordList))
	for i := 0; i < len(keywordList); i++ {
		if len(keywordList[i]) > 0 {
			availableList = append(availableList, keywordList[i])
		}
	}
	if len(availableList) == 0 {
		return shares, 0
	}

	dbChain := DB
	dbChain = dbChain.Where("password = ? and remain_downloads <> 0 and (expires is NULL or expires > ?) and source_name like ?", "", time.Now(), "%"+strings.Join(availableList, "%")+"%")

	// 计算总数用于分页
	dbChain.Model(&Share{}).Count(&total)

	// 查询记录
	dbChain.Limit(pageSize).Offset((page - 1) * pageSize).Order(order).Find(&shares)
	return shares, total
}
