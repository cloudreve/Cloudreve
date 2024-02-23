package payment

import (
	"encoding/json"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"strconv"
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
	cache.Deletes([]string{strconv.FormatUint(uint64(user.ID), 10)}, "pack_size_")
	return nil
}

func checkGroupUpgrade(user *model.User, groupInfo *serializer.GroupProducts) error {
	if user.Group.ID == 1 {
		return ErrAdminFulfillGroup
	}

	// 检查用户是否已有未过期用户
	if user.PreviousGroupID != 0 && user.GroupID != groupInfo.GroupID {
		return ErrGroupConflict
	}

	// 用户组不能相同
	if user.GroupID == groupInfo.GroupID && user.PreviousGroupID == 0 {
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
	if user.PreviousGroupID != 0 {
		expires = user.GroupExpires.Add(time.Duration(groupInfo.Time*int64(num)) * time.Second)
	}

	if err := user.UpgradeGroup(groupInfo.GroupID, &expires); err != nil {
		return ErrUpgradeGroup.WithError(err)
	}

	return nil
}

// GiveScore 积分充值
func GiveScore(user *model.User, num int) error {
	user.AddScore(num)
	return nil
}

// GiveProduct “发货”
func GiveProduct(user *model.User, pack *serializer.PackProduct, group *serializer.GroupProducts, num int) error {
	if pack != nil {
		return GivePack(user, pack, num)
	} else if group != nil {
		return GiveGroup(user, group, num)
	} else {
		return GiveScore(user, num)
	}
}

// OrderPaid 订单已支付处理
func OrderPaid(orderNo string) error {
	order, err := model.GetOrderByNo(orderNo)
	if err != nil || order.Status == model.OrderPaid {
		return ErrOrderNotFound.WithError(err)
	}

	// 更新订单状态为 已支付
	order.UpdateStatus(model.OrderPaid)

	user, err := model.GetActiveUserByID(order.UserID)
	if err != nil {
		return serializer.NewError(serializer.CodeUserNotFound, "", err)
	}

	// 查询商品
	options := model.GetSettingByNames("pack_data", "group_sell_data")

	var (
		packs  []serializer.PackProduct
		groups []serializer.GroupProducts
	)
	if err := json.Unmarshal([]byte(options["pack_data"]), &packs); err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(options["group_sell_data"]), &groups); err != nil {
		return err
	}

	// 查找要购买的商品
	var (
		pack  *serializer.PackProduct
		group *serializer.GroupProducts
	)
	if order.Type == model.GroupOrderType {
		for _, v := range groups {
			if v.ID == order.ProductID {
				group = &v
				break
			}
		}
	} else if order.Type == model.PackOrderType {
		for _, v := range packs {
			if v.ID == order.ProductID {
				pack = &v
				break
			}
		}
	}

	// "发货"
	return GiveProduct(&user, pack, group, order.Num)

}
