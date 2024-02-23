package vas

import (
	"encoding/json"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/payment"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/gin-gonic/gin"
)

// CreateOrderService 创建订单服务
type CreateOrderService struct {
	Action string `json:"action" binding:"required,eq=group|eq=pack|eq=score"`
	Method string `json:"method" binding:"required,eq=alipay|eq=score|eq=payjs|eq=wechat|eq=custom"`
	ID     int64  `json:"id" binding:"required"`
	Num    int    `json:"num" binding:"required,min=1"`
}

// RedeemService 兑换服务
type RedeemService struct {
	Code string `uri:"code" binding:"required,max=64"`
}

// OrderService 订单查询
type OrderService struct {
	ID string `uri:"id" binding:"required"`
}

// Status 查询订单状态
func (service *OrderService) Status(c *gin.Context, user *model.User) serializer.Response {
	order, _ := model.GetOrderByNo(service.ID)
	if order == nil || order.UserID != user.ID {
		return serializer.Err(serializer.CodeNotFound, "", nil)
	}

	return serializer.Response{Data: order.Status}
}

// Redeem 开始兑换
func (service *RedeemService) Redeem(c *gin.Context, user *model.User) serializer.Response {
	redeem, err := model.GetAvailableRedeem(service.Code)
	if err != nil {
		return serializer.Err(serializer.CodeInvalidGiftCode, "", err)
	}

	// 取得当前商品信息
	packs, groups, err := decodeProductInfo()
	if err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "Failed to parse product settings", err)
	}

	// 查找要购买的商品
	var (
		pack  *serializer.PackProduct
		group *serializer.GroupProducts
	)
	if redeem.Type == model.GroupOrderType {
		for _, v := range groups {
			if v.ID == redeem.ProductID {
				group = &v
				break
			}
		}

		if group == nil {
			return serializer.Err(serializer.CodeNotFound, "", err)
		}

	} else if redeem.Type == model.PackOrderType {
		for _, v := range packs {
			if v.ID == redeem.ProductID {
				pack = &v
				break
			}
		}

		if pack == nil {
			return serializer.Err(serializer.CodeNotFound, "", err)
		}

	}

	err = payment.GiveProduct(user, pack, group, redeem.Num)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, "Redeem failed", err)
	}

	redeem.Use()

	return serializer.Response{}

}

// Query 检查兑换码信息
func (service *RedeemService) Query(c *gin.Context) serializer.Response {
	redeem, err := model.GetAvailableRedeem(service.Code)
	if err != nil {
		return serializer.Err(serializer.CodeInvalidGiftCode, "", err)
	}

	var (
		name        = "积分"
		productTime int64
	)
	if redeem.Type != model.ScoreOrderType {
		packs, groups, err := decodeProductInfo()
		if err != nil {
			return serializer.Err(serializer.CodeInternalSetting, "Failed to parse product settings", err)
		}
		if redeem.Type == model.GroupOrderType {
			for _, v := range groups {
				if v.ID == redeem.ProductID {
					name = v.Name
					productTime = v.Time
					break
				}
			}
		} else {
			for _, v := range packs {
				if v.ID == redeem.ProductID {
					name = v.Name
					productTime = v.Time
					break
				}
			}
		}

		if name == "积分" {
			return serializer.Err(serializer.CodeNotFound, "", err)
		}

	}

	return serializer.Response{
		Data: struct {
			Name string `json:"name"`
			Type int    `json:"type"`
			Num  int    `json:"num"`
			Time int64  `json:"time"`
		}{
			name, redeem.Type, redeem.Num, productTime,
		},
	}
}

// Create 创建新订单
func (service *CreateOrderService) Create(c *gin.Context, user *model.User) serializer.Response {
	// 取得当前商品信息
	packs, groups, err := decodeProductInfo()
	if err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "Failed to parse product list", err)
	}

	// 查找要购买的商品
	var (
		pack  *serializer.PackProduct
		group *serializer.GroupProducts
	)
	if service.Action == "group" {
		for _, v := range groups {
			if v.ID == service.ID {
				group = &v
				break
			}
		}
	} else if service.Action == "pack" {
		for _, v := range packs {
			if v.ID == service.ID {
				pack = &v
				break
			}
		}
	}

	// 购买积分
	if pack == nil && group == nil {
		if service.Method == "score" {
			return serializer.ParamErr("Payment method not supported", nil)
		}
	}

	// 创建订单
	res, err := payment.NewOrder(pack, group, service.Num, service.Method, user)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	return serializer.Response{Data: res}

}

// Products 获取商品信息
func (service *GeneralVASService) Products(c *gin.Context, user *model.User) serializer.Response {
	options := model.GetSettingByNames(
		"wechat_enabled",
		"alipay_enabled",
		"payjs_enabled",
		"payjs_enabled",
		"custom_payment_enabled",
		"custom_payment_name",
	)
	scorePrice := model.GetIntSetting("score_price", 0)
	packs, groups, err := decodeProductInfo()
	if err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "Failed to parse product list", err)
	}

	return serializer.BuildProductResponse(
		groups,
		packs,
		model.IsTrueVal(options["wechat_enabled"]),
		model.IsTrueVal(options["alipay_enabled"]),
		model.IsTrueVal(options["payjs_enabled"]),
		model.IsTrueVal(options["custom_payment_enabled"]),
		options["custom_payment_name"],
		scorePrice,
	)
}

func decodeProductInfo() ([]serializer.PackProduct, []serializer.GroupProducts, error) {
	options := model.GetSettingByNames("pack_data", "group_sell_data", "alipay_enabled", "payjs_enabled")

	var (
		packs  []serializer.PackProduct
		groups []serializer.GroupProducts
	)
	if err := json.Unmarshal([]byte(options["pack_data"]), &packs); err != nil {
		return nil, nil, err
	}
	if err := json.Unmarshal([]byte(options["group_sell_data"]), &groups); err != nil {
		return nil, nil, err
	}

	return packs, groups, nil
}
