package payment

import (
	"fmt"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/serializer"
	alipay "github.com/smartwalle/alipay/v3"
	"net/url"
)

// Alipay 支付宝当面付支付处理
type Alipay struct {
	Client *alipay.Client
}

// Create 创建订单
func (pay *Alipay) Create(order *model.Order, pack *serializer.PackProduct, group *serializer.GroupProducts, user *model.User) (*OrderCreateRes, error) {
	gateway, _ := url.Parse("/api/v3/callback/alipay")
	var p = alipay.TradePreCreate{
		Trade: alipay.Trade{
			NotifyURL:   model.GetSiteURL().ResolveReference(gateway).String(),
			Subject:     order.Name,
			OutTradeNo:  order.OrderNo,
			TotalAmount: fmt.Sprintf("%.2f", float64(order.Price*order.Num)/100),
		},
	}

	if _, err := order.Create(); err != nil {
		return nil, ErrInsertOrder.WithError(err)
	}

	res, err := pay.Client.TradePreCreate(p)
	if err != nil {
		return nil, ErrIssueOrder.WithError(err)
	}

	return &OrderCreateRes{
		Payment: true,
		QRCode:  res.Content.QRCode,
		ID:      order.OrderNo,
	}, nil
}
