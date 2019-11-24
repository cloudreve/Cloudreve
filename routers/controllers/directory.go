package controllers

import (
	"github.com/HFO4/cloudreve/service/explorer"
	"github.com/gin-gonic/gin"
)

// CreateDirectory 创建目录
func CreateDirectory(c *gin.Context) {
	var service explorer.DirectoryCreateService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.CreateDirectory(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}
