package controllers

import (
	"context"

	"github.com/cloudreve/Cloudreve/v3/service/explorer"
	"github.com/gin-gonic/gin"
)

// Delete 删除文件或目录
func Delete(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.ItemIDService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Delete(ctx, c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// Move 移动文件或目录
func Move(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.ItemMoveService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Move(ctx, c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// Copy 复制文件或目录
func Copy(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.ItemMoveService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Copy(ctx, c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// Rename 重命名文件或目录
func Rename(c *gin.Context) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var service explorer.ItemRenameService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Rename(ctx, c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}
