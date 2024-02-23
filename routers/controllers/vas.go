package controllers

import (
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/payment"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/cloudreve/Cloudreve/v3/service/vas"
	"github.com/gin-gonic/gin"
	"github.com/iGoogle-ink/gopay"
	"github.com/iGoogle-ink/gopay/wechat/v3"
	"github.com/qingwg/payjs/notify"
	"github.com/smartwalle/alipay/v3"
	"net/http"
)

// GetQuota 获取容量配额信息
func GetQuota(c *gin.Context) {
	var service vas.GeneralVASService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.Quota(c, CurrentUser(c))
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// GetProduct 获取商品信息
func GetProduct(c *gin.Context) {
	var service vas.GeneralVASService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.Products(c, CurrentUser(c))
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// NewOrder 新建支付订单
func NewOrder(c *gin.Context) {
	var service vas.CreateOrderService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Create(c, CurrentUser(c))
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// OrderStatus 查询订单状态
func OrderStatus(c *gin.Context) {
	var service vas.OrderService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.Status(c, CurrentUser(c))
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// GetRedeemInfo 获取兑换码信息
func GetRedeemInfo(c *gin.Context) {
	var service vas.RedeemService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.Query(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// DoRedeem 获取兑换码信息
func DoRedeem(c *gin.Context) {
	var service vas.RedeemService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.Redeem(c, CurrentUser(c))
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// AlipayCallback 支付宝回调
func AlipayCallback(c *gin.Context) {
	pay, err := payment.NewPaymentInstance("alipay")
	if err != nil {
		util.Log().Debug("[Alipay callback] Failed to create alipay client, %s", err)
		c.Status(400)
		return
	}

	res, err := pay.(*payment.Alipay).Client.GetTradeNotification(c.Request)
	if err != nil {
		util.Log().Debug("[Alipay callback] Failed to validate callback request, %s", err)
		c.Status(403)
		return
	}

	if res != nil && res.TradeStatus == "TRADE_SUCCESS" {
		// 支付成功
		if err := payment.OrderPaid(res.OutTradeNo); err != nil {
			util.Log().Debug("[Alipay callback] Failed to process payment, %s", err)
		}
	}

	// 确认收到通知消息
	alipay.AckNotification(c.Writer)
}

// WechatCallback 微信扫码支付回调
func WechatCallback(c *gin.Context) {
	pay, err := payment.NewPaymentInstance("wechat")
	if err != nil {
		util.Log().Debug("[Wechat pay callback] Failed to create alipay client, %s", err)
		c.JSON(500, &wechat.V3NotifyRsp{Code: gopay.FAIL, Message: "Failed to create alipay client"})
		return
	}

	notifyReq, err := wechat.V3ParseNotify(c.Request)
	if err != nil {
		util.Log().Debug("[Wechat pay callback] Failed to parse callback content, %s", err)
		c.JSON(500, &wechat.V3NotifyRsp{Code: gopay.FAIL, Message: "Failed to parse callback content"})
		return
	}

	err = notifyReq.VerifySign(pay.(*payment.Wechat).GetPlatformCert())
	if err != nil {
		util.Log().Debug("[Wechat pay callback] Failed to verify callback signature, %s", err)
		c.JSON(403, &wechat.V3NotifyRsp{Code: gopay.FAIL, Message: "Failed to verify callback signature"})
		return
	}

	// 解密回调正文
	result, err := notifyReq.DecryptCipherText(pay.(*payment.Wechat).ApiV3Key)
	if result != nil && result.TradeState == "SUCCESS" {
		// 支付成功
		if err := payment.OrderPaid(result.OutTradeNo); err != nil {
			util.Log().Debug("[Wechat pay callback] Failed to process payment, %s", err)
		}
	}

	// 确认收到通知消息
	c.JSON(http.StatusOK, &wechat.V3NotifyRsp{Code: gopay.SUCCESS, Message: "Success"})
}

// PayJSCallback PayJS回调
func PayJSCallback(c *gin.Context) {
	pay, err := payment.NewPaymentInstance("payjs")
	if err != nil {
		util.Log().Debug("[PayJS callback] Failed to initialize payment client, %s", err)
		c.Status(400)
		return
	}

	payNotify := pay.(*payment.PayJSClient).Client.GetNotify(c.Request, c.Writer)

	//设置接收消息的处理方法
	payNotify.SetMessageHandler(func(msg notify.Message) {
		if err := payment.OrderPaid(msg.OutTradeNo); err != nil {
			util.Log().Debug("[PayJS callback] Failed to process payment, %s", err)
		}
	})

	//处理消息接收以及回复
	err = payNotify.Serve()
	if err != nil {
		util.Log().Debug("[PayJS callback] Failed to process payment, %s", err)
		return
	}

	//发送回复的消息
	payNotify.SendResponseMsg()

}

// QQCallback QQ互联回调
func QQCallback(c *gin.Context) {
	var service vas.QQCallbackService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Callback(c, CurrentUser(c))
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// CustomCallback PayJS回调
func CustomCallback(c *gin.Context) {
	orderNo := c.Param("orderno")
	sessionID := c.Param("id")
	sessionRaw, exist := cache.Get(payment.CallbackSessionPrefix + sessionID)
	if !exist {
		util.Log().Debug("[Custom callback] Failed to process payment, session not found")
		c.JSON(200, serializer.Err(serializer.CodeNotFound, "session not found", nil))
		return
	}

	expectedID := sessionRaw.(string)
	if expectedID != orderNo {
		util.Log().Debug("[Custom callback] Failed to process payment, session mismatch")
		c.JSON(200, serializer.Err(serializer.CodeInternalSetting, "session mismatch", nil))
		return
	}

	cache.Deletes([]string{sessionID}, payment.CallbackSessionPrefix)

	if err := payment.OrderPaid(orderNo); err != nil {
		c.JSON(200, serializer.Err(serializer.CodeInternalSetting, "failed to fulfill payment", err))
		util.Log().Debug("[Custom callback] Failed to process payment, %s", err)
		return
	}

	c.JSON(200, serializer.Response{})
}
