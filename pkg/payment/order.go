package payment

import (
	"fmt"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/serializer"
)

var (
	// ErrUnknownPaymentMethod 未知支付方式
	ErrUnknownPaymentMethod = serializer.NewError(serializer.CodeNotFound, "未知支付方式", nil)
	// ErrUnsupportedPaymentMethod 未知支付方式
	ErrUnsupportedPaymentMethod = serializer.NewError(serializer.CodeNotFound, "此订单不支持此支付方式", nil)
	// ErrInsertOrder 无法插入订单记录
	ErrInsertOrder = serializer.NewError(serializer.CodeDBError, "无法插入订单记录", nil)
	// ErrScoreNotEnough 积分不足
	ErrScoreNotEnough = serializer.NewError(serializer.CodeNoPermissionErr, "积分不足", nil)
	// ErrCreateStoragePack 无法创建容量包
	ErrCreateStoragePack = serializer.NewError(serializer.CodeNoPermissionErr, "无法创建容量包", nil)
	// ErrGroupConflict 用户组冲突
	ErrGroupConflict = serializer.NewError(serializer.CodeNoPermissionErr, "当前用户组仍未过期，请前往个人设置手动解约后继续", nil)
	// ErrGroupInvalid 用户组冲突
	ErrGroupInvalid = serializer.NewError(serializer.CodeNoPermissionErr, "用户组不可用", nil)
	// ErrUpgradeGroup 用户组冲突
	ErrUpgradeGroup = serializer.NewError(serializer.CodeDBError, "无法升级用户组", nil)
)

// Pay 支付处理接口
type Pay interface {
	Create(order *model.Order, pack *serializer.PackProduct, group *serializer.GroupProducts, user *model.User) (*OrderCreateRes, error)
}

// OrderCreateRes 订单创建结果
type OrderCreateRes struct {
	Payment bool `json:"payment"` // 是否需要支付
}

// NewPaymentInstance 获取新的支付实例
func NewPaymentInstance(method string) (Pay, error) {
	if method == "score" {
		return &ScorePayment{}, nil
	}

	return nil, ErrUnknownPaymentMethod
}

// NewOrder 创建新订单
func NewOrder(pack *serializer.PackProduct, group *serializer.GroupProducts, num int, method string, user *model.User) (*OrderCreateRes, error) {
	// 获取支付实例
	pay, err := NewPaymentInstance(method)
	if err != nil {
		return nil, err
	}

	var (
		orderType int
		productID int64
		title     string
		price     int
	)
	if pack != nil {
		orderType = model.PackOrderType
		productID = pack.ID
		title = pack.Name
		price = pack.Price
	} else if group != nil {
		orderType = model.GroupOrderType
		productID = group.ID
		title = group.Name
		price = group.Price
	} else {
		orderType = model.ScoreOrderType
		productID = 0
		title = fmt.Sprintf("%d 积分", num)
		price = model.GetIntSetting("score_price", 1)
	}

	// 创建订单记录
	order := &model.Order{
		UserID:    user.ID,
		Type:      orderType,
		Method:    method,
		ProductID: productID,
		Num:       num,
		Name:      fmt.Sprintf("%s - %s", model.GetSettingByName("siteName"), title),
		Price:     price,
		Status:    model.OrderUnpaid,
	}

	return pay.Create(order, pack, group, user)
}
