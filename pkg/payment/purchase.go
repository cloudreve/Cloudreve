package payment

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"time"
)

// GivePack 创建容量包
func GivePack(user *model.User, packInfo *serializer.PackProduct, num int) error {
	timeNow := time.Now()
	expires := timeNow.Add(time.Duration(packInfo.Time*int64(num)) * time.Second)
	pack := model.StoragePack{
		Name:        packInfo.Name,
		UserID:      user.ID,
		ActiveTime:  &timeNow,
		ExpiredTime: &expires,
		Size:        packInfo.Size,
	}
	if _, err := pack.Create(); err != nil {
		return ErrCreateStoragePack.WithError(err)
	}
	return nil
}

func checkGroupUpgrade(user *model.User, groupInfo *serializer.GroupProducts) error {
	// 检查用户是否已有未过期用户
	if user.PreviousGroupID != 0 {
		return ErrGroupConflict
	}

	// 用户组不能相同
	if user.GroupID == groupInfo.GroupID {
		return ErrGroupInvalid
	}

	return nil
}

// GiveGroup 升级用户组
func GiveGroup(user *model.User, groupInfo *serializer.GroupProducts, num int) error {
	if err := checkGroupUpgrade(user, groupInfo); err != nil {
		return err
	}

	timeNow := time.Now()
	expires := timeNow.Add(time.Duration(groupInfo.Time*int64(num)) * time.Second)

	if err := user.UpgradeGroup(groupInfo.GroupID, &expires); err != nil {
		return ErrUpgradeGroup.WithError(err)
	}

	return nil
}

// GiveProduct “发货”
func GiveProduct(user *model.User, pack *serializer.PackProduct, group *serializer.GroupProducts, num int) error {
	if pack != nil {
		return GivePack(user, pack, num)
	} else if group != nil {
		return GiveGroup(user, group, num)
	}
	return nil
}
