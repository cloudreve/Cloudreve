package payment

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/serializer"
)

// ScorePayment 积分支付处理
type ScorePayment struct {
}

// Create 创建新订单
func (pay *ScorePayment) Create(order *model.Order, pack *serializer.PackProduct, group *serializer.GroupProducts, user *model.User) (*OrderCreateRes, error) {
	if pack != nil {
		order.Price = pack.Score
	} else {
		order.Price = group.Score
	}

	// 检查此订单是否可用积分支付
	if order.Price == 0 {
		return nil, ErrUnsupportedPaymentMethod
	}

	// 创建订单记录
	order.Status = model.OrderPaid
	if _, err := order.Create(); err != nil {
		return nil, ErrInsertOrder.WithError(err)
	}

	return &OrderCreateRes{
		Payment: false,
	}, nil
}
