package payment

import (
	"fmt"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/iGoogle-ink/gopay/wechat/v3"
	"github.com/qingwg/payjs"
	"github.com/smartwalle/alipay/v3"
	"math/rand"
	"net/url"
	"time"
)

var (
	// ErrUnknownPaymentMethod 未知支付方式
	ErrUnknownPaymentMethod = serializer.NewError(serializer.CodeInternalSetting, "Unknown payment method", nil)
	// ErrUnsupportedPaymentMethod 未知支付方式
	ErrUnsupportedPaymentMethod = serializer.NewError(serializer.CodeInternalSetting, "This order cannot be paid with this method", nil)
	// ErrInsertOrder 无法插入订单记录
	ErrInsertOrder = serializer.NewError(serializer.CodeDBError, "Failed to insert order record", nil)
	// ErrScoreNotEnough 积分不足
	ErrScoreNotEnough = serializer.NewError(serializer.CodeInsufficientCredit, "", nil)
	// ErrCreateStoragePack 无法创建容量包
	ErrCreateStoragePack = serializer.NewError(serializer.CodeDBError, "Failed to create storage pack record", nil)
	// ErrGroupConflict 用户组冲突
	ErrGroupConflict = serializer.NewError(serializer.CodeGroupConflict, "", nil)
	// ErrGroupInvalid 用户组冲突
	ErrGroupInvalid = serializer.NewError(serializer.CodeGroupInvalid, "", nil)
	// ErrAdminFulfillGroup 管理员无法购买用户组
	ErrAdminFulfillGroup = serializer.NewError(serializer.CodeFulfillAdminGroup, "", nil)
	// ErrUpgradeGroup 用户组冲突
	ErrUpgradeGroup = serializer.NewError(serializer.CodeDBError, "Failed to update user's group", nil)
	// ErrUInitPayment 无法初始化支付实例
	ErrUInitPayment = serializer.NewError(serializer.CodeInternalSetting, "Failed to initialize payment client", nil)
	// ErrIssueOrder 订单接口请求失败
	ErrIssueOrder = serializer.NewError(serializer.CodeInternalSetting, "Failed to create order", nil)
	// ErrOrderNotFound 订单不存在
	ErrOrderNotFound = serializer.NewError(serializer.CodeNotFound, "", nil)
)

// Pay 支付处理接口
type Pay interface {
	Create(order *model.Order, pack *serializer.PackProduct, group *serializer.GroupProducts, user *model.User) (*OrderCreateRes, error)
}

// OrderCreateRes 订单创建结果
type OrderCreateRes struct {
	Payment bool   `json:"payment"`           // 是否需要支付
	ID      string `json:"id,omitempty"`      // 订单号
	QRCode  string `json:"qr_code,omitempty"` // 支付二维码指向的地址
}

// NewPaymentInstance 获取新的支付实例
func NewPaymentInstance(method string) (Pay, error) {
	switch method {
	case "score":
		return &ScorePayment{}, nil
	case "alipay":
		options := model.GetSettingByNames("alipay_enabled", "appid", "appkey", "shopid")
		if options["alipay_enabled"] != "1" {
			return nil, ErrUnknownPaymentMethod
		}

		// 初始化支付宝客户端
		var client, err = alipay.New(options["appid"], options["appkey"], true)
		if err != nil {
			return nil, ErrUInitPayment.WithError(err)
		}

		// 加载支付宝公钥
		err = client.LoadAliPayPublicKey(options["shopid"])
		if err != nil {
			return nil, ErrUInitPayment.WithError(err)
		}

		return &Alipay{Client: client}, nil
	case "payjs":
		options := model.GetSettingByNames("payjs_enabled", "payjs_secret", "payjs_id")
		if options["payjs_enabled"] != "1" {
			return nil, ErrUnknownPaymentMethod
		}

		callback, _ := url.Parse("/api/v3/callback/payjs")
		payjsConfig := &payjs.Config{
			Key:       options["payjs_secret"],
			MchID:     options["payjs_id"],
			NotifyUrl: model.GetSiteURL().ResolveReference(callback).String(),
		}

		return &PayJSClient{Client: payjs.New(payjsConfig)}, nil
	case "wechat":
		options := model.GetSettingByNames("wechat_enabled", "wechat_appid", "wechat_mchid", "wechat_serial_no", "wechat_api_key", "wechat_pk_content")
		if options["wechat_enabled"] != "1" {
			return nil, ErrUnknownPaymentMethod
		}
		client, err := wechat.NewClientV3(options["wechat_appid"], options["wechat_mchid"], options["wechat_serial_no"], options["wechat_api_key"], options["wechat_pk_content"])
		if err != nil {
			return nil, ErrUInitPayment.WithError(err)
		}

		return &Wechat{Client: client, ApiV3Key: options["wechat_api_key"]}, nil
	case "custom":
		options := model.GetSettingByNames("custom_payment_enabled", "custom_payment_endpoint", "custom_payment_secret")
		if !model.IsTrueVal(options["custom_payment_enabled"]) {
			return nil, ErrUnknownPaymentMethod
		}

		return newCustomClient(options["custom_payment_endpoint"], options["custom_payment_secret"]), nil
	default:
		return nil, ErrUnknownPaymentMethod
	}
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
		if err := checkGroupUpgrade(user, group); err != nil {
			return nil, err
		}

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
		OrderNo:   orderID(),
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

func orderID() string {
	return fmt.Sprintf("%s%d",
		time.Now().Format("20060102150405"),
		100000+rand.Intn(900000),
	)
}
