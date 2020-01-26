package controllers

import (
	"github.com/HFO4/cloudreve/service/share"
	"github.com/gin-gonic/gin"
)

// CreateShare 创建分享
func CreateShare(c *gin.Context) {
	var service share.ShareCreateService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Create(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}
