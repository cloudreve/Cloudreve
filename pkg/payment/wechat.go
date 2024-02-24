package payment

import (
	"errors"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/iGoogle-ink/gopay"
	"github.com/iGoogle-ink/gopay/wechat/v3"
	"net/url"
	"time"
)

// Wechat 微信扫码支付接口
type Wechat struct {
	Client   *wechat.ClientV3
	ApiV3Key string
}

// Create 创建订单
func (pay *Wechat) Create(order *model.Order, pack *serializer.PackProduct, group *serializer.GroupProducts, user *model.User) (*OrderCreateRes, error) {
	gateway, _ := url.Parse("/api/v3/callback/wechat")
	bm := make(gopay.BodyMap)
	bm.
		Set("description", order.Name).
		Set("out_trade_no", order.OrderNo).
		Set("notify_url", model.GetSiteURL().ResolveReference(gateway).String()).
		SetBodyMap("amount", func(bm gopay.BodyMap) {
			bm.Set("total", int64(order.Price*order.Num)).
				Set("currency", "CNY")
		})

	wxRsp, err := pay.Client.V3TransactionNative(bm)
	if err != nil {
		return nil, ErrIssueOrder.WithError(err)
	}

	if wxRsp.Code == wechat.Success {
		if _, err := order.Create(); err != nil {
			return nil, ErrInsertOrder.WithError(err)
		}

		return &OrderCreateRes{
			Payment: true,
			QRCode:  wxRsp.Response.CodeUrl,
			ID:      order.OrderNo,
		}, nil
	}

	return nil, ErrIssueOrder.WithError(errors.New(wxRsp.Error))
}

// GetPlatformCert 获取微信平台证书
func (pay *Wechat) GetPlatformCert() string {
	if cert, ok := cache.Get("wechat_platform_cert"); ok {
		return cert.(string)
	}

	res, err := pay.Client.GetPlatformCerts()
	if err == nil {
		// 使用反馈证书中启用时间较晚的
		var (
			currentLatest *time.Time
			currentCert   string
		)
		for _, cert := range res.Certs {
			effectiveTime, err := time.Parse("2006-01-02T15:04:05-0700", cert.EffectiveTime)
			if err != nil {
				if currentLatest == nil {
					currentLatest = &effectiveTime
					currentCert = cert.PublicKey
					continue
				}
				if currentLatest.Before(effectiveTime) {
					currentLatest = &effectiveTime
					currentCert = cert.PublicKey
				}
			}
		}

		cache.Set("wechat_platform_cert", currentCert, 3600*10)
		return currentCert
	}

	util.Log().Debug("Failed to get Wechat Pay platform certificate: %s", err)
	return ""
}
