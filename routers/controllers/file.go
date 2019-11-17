package controllers

import (
	"context"
	"github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/service/file"
	"github.com/gin-gonic/gin"
)

// FileUpload 本地策略文件上传
func FileUpload(c *gin.Context) {

	// 非本地策略时拒绝上传
	if user, ok := c.Get("user"); ok && user.(*model.User).Policy.Type != "local" {
		c.JSON(200, serializer.Err(serializer.CodePolicyNotAllowed, "当前上传策略无法使用", nil))
		return
	}

	// 建立上下文
	ctx, cancel := context.WithCancel(context.Background())

	var service file.UploadService
	defer cancel()

	if err := c.ShouldBind(&service); err == nil {
		res := service.Upload(ctx, c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}
