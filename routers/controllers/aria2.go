package controllers

import (
	"github.com/HFO4/cloudreve/service/aria2"
	"github.com/gin-gonic/gin"
)

// AddAria2URL 添加离线下载URL
func AddAria2URL(c *gin.Context) {
	var addService aria2.AddURLService
	if err := c.ShouldBindJSON(&addService); err == nil {
		res := addService.Add(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}
