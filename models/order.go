package model

import (
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/jinzhu/gorm"
)

const (
	// PackOrderType 容量包订单
	PackOrderType = iota
	// GroupOrderType 用户组订单
	GroupOrderType
	// ScoreOrderType 积分充值订单
	ScoreOrderType
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
	OrderNo   string `gorm:"index:order_number"` // 商户自定义订单编号
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
		util.Log().Warning("Failed to insert order record: %s", err)
		return 0, err
	}
	return order.ID, nil
}

// UpdateStatus 更新订单状态
func (order *Order) UpdateStatus(status int) {
	DB.Model(order).Update("status", status)
}

// GetOrderByNo 根据商户订单号查询订单
func GetOrderByNo(id string) (*Order, error) {
	var order Order
	err := DB.Where("order_no = ?", id).First(&order).Error
	return &order, err
}
