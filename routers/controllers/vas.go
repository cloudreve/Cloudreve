package controllers

import (
	"github.com/HFO4/cloudreve/service/vas"
	"github.com/gin-gonic/gin"
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
