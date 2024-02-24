package payment

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/qingwg/payjs"
)

// PayJSClient PayJS支付处理
type PayJSClient struct {
	Client *payjs.PayJS
}

// Create 创建订单
func (pay *PayJSClient) Create(order *model.Order, pack *serializer.PackProduct, group *serializer.GroupProducts, user *model.User) (*OrderCreateRes, error) {
	if _, err := order.Create(); err != nil {
		return nil, ErrInsertOrder.WithError(err)
	}

	PayNative := pay.Client.GetNative()
	res, err := PayNative.Create(int64(order.Price*order.Num), order.Name, order.OrderNo, "", "")
	if err != nil {
		return nil, ErrIssueOrder.WithError(err)
	}

	return &OrderCreateRes{
		Payment: true,
		QRCode:  res.CodeUrl,
		ID:      order.OrderNo,
	}, nil
}
