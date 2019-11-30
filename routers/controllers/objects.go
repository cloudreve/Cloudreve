package controllers

import (
	"context"
	"github.com/HFO4/cloudreve/service/explorer"
	"github.com/gin-gonic/gin"
)

// Delete 删除文件或目录
func Delete(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.ItemService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Delete(ctx, c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}
