package model

import (
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/jinzhu/gorm"
)

const (
	// PackOrderType 容量包订单
	PackOrderType = iota
	// GroupOrderType 用户组订单
	GroupOrderType
)

const (
	// OrderUnpaid 未支付
	OrderUnpaid = iota
	// OrderPaid 已支付
	OrderPaid
	// OrderCanceled 已取消
	OrderCanceled
)

// Order 交易订单
type Order struct {
	gorm.Model
	UserID    uint   // 创建者ID
	OrderNo   string // 商户自定义订单编号
	Type      int    // 订单类型
	Method    string // 支付类型
	ProductID int64  // 商品ID
	Num       int    // 商品数量
	Name      string // 订单标题
	Price     int    // 商品单价
	Status    int    // 订单状态
}

// Create 创建订单记录
func (order *Order) Create() (uint, error) {
	if err := DB.Create(order).Error; err != nil {
		util.Log().Warning("无法插入离线下载记录, %s", err)
		return 0, err
	}
	return order.ID, nil
}
