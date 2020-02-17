package vas

import (
	"encoding/json"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/payment"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/gin-gonic/gin"
)

// GeneralVASService 通用增值服务
type GeneralVASService struct {
}

// CreateOrderService 创建订单服务
type CreateOrderService struct {
	Action string `json:"action" binding:"required,eq=group|eq=pack|eq=score"`
	Method string `json:"method" binding:"required,eq=alipay|eq=score|eq=payjs"`
	ID     int64  `json:"id" binding:"required"`
	Num    int    `json:"num" binding:"required,min=1"`
}

// Create 创建新订单
func (service *CreateOrderService) Create(c *gin.Context, user *model.User) serializer.Response {
	// 取得当前商品信息
	packs, groups, err := decodeProductInfo()
	if err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "无法解析商品设置", err)
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
			return serializer.Err(serializer.CodeNotFound, "不支持此支付方式", nil)
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
	options := model.GetSettingByNames("alipay_enabled", "payjs_enabled", "score_price")
	scorePrice := model.GetIntSetting("score_price", 0)
	packs, groups, err := decodeProductInfo()
	if err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "无法解析商品设置", err)
	}

	return serializer.BuildProductResponse(groups, packs, options["alipay_enabled"] == "1", options["payjs_enabled"] == "1", scorePrice)
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

// Quota 获取容量配额信息
func (service *GeneralVASService) Quota(c *gin.Context, user *model.User) serializer.Response {
	packs := user.GetAvailableStoragePacks()
	return serializer.BuildUserQuotaResponse(user, packs)
}
