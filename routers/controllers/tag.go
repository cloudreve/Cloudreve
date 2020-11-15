package controllers

import (
	"github.com/cloudreve/Cloudreve/v3/service/explorer"
	"github.com/gin-gonic/gin"
)

// CreateFilterTag 创建文件分类标签
func CreateFilterTag(c *gin.Context) {
	var service explorer.FilterTagCreateService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Create(c, CurrentUser(c))
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// CreateLinkTag 创建目录快捷方式标签
func CreateLinkTag(c *gin.Context) {
	var service explorer.LinkTagCreateService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Create(c, CurrentUser(c))
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// DeleteTag 删除标签
func DeleteTag(c *gin.Context) {
	var service explorer.TagService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.Delete(c, CurrentUser(c))
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}
