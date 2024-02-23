package model

import (
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/jinzhu/gorm"
	"strconv"
	"time"
)

// StoragePack 容量包模型
type StoragePack struct {
	// 表字段
	gorm.Model
	Name        string
	UserID      uint
	ActiveTime  *time.Time
	ExpiredTime *time.Time `gorm:"index:expired"`
	Size        uint64
}

// Create 创建容量包
func (pack *StoragePack) Create() (uint, error) {
	if err := DB.Create(pack).Error; err != nil {
		util.Log().Warning("Failed to insert storage pack record: %s", err)
		return 0, err
	}
	return pack.ID, nil
}

// GetAvailablePackSize 返回给定用户当前可用的容量包总容量
func (user *User) GetAvailablePackSize() uint64 {
	var (
		total       uint64
		firstExpire *time.Time
		timeNow     = time.Now()
		ttl         int64
	)

	// 尝试从缓存中读取
	cacheKey := "pack_size_" + strconv.FormatUint(uint64(user.ID), 10)
	if total, ok := cache.Get(cacheKey); ok {
		return total.(uint64)
	}

	// 查找所有有效容量包
	packs := user.GetAvailableStoragePacks()

	// 计算总容量, 并找到其中最早的过期时间
	for _, v := range packs {
		total += v.Size
		if firstExpire == nil {
			firstExpire = v.ExpiredTime
			continue
		}
		if v.ExpiredTime != nil && firstExpire.After(*v.ExpiredTime) {
			firstExpire = v.ExpiredTime
		}
	}

	// 用最早的过期时间计算缓存TTL，并写入缓存
	if firstExpire != nil {
		ttl = firstExpire.Unix() - timeNow.Unix()
		if ttl > 0 {
			_ = cache.Set(cacheKey, total, int(ttl))
		}
	}

	return total
}

// GetAvailableStoragePacks 返回用户可用的容量包
func (user *User) GetAvailableStoragePacks() []StoragePack {
	var packs []StoragePack
	timeNow := time.Now()
	// 查找所有有效容量包
	DB.Where("expired_time > ? AND user_id = ?", timeNow, user.ID).Find(&packs)
	return packs
}

// GetExpiredStoragePack 获取已过期的容量包
func GetExpiredStoragePack() []StoragePack {
	var packs []StoragePack
	DB.Where("expired_time < ?", time.Now()).Find(&packs)
	return packs
}

// Delete 删除容量包
func (pack *StoragePack) Delete() error {
	return DB.Delete(&pack).Error
}
