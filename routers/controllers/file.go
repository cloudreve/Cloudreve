package controllers

import (
	"cloudreve/models"
	"cloudreve/pkg/serializer"
	"github.com/gin-gonic/gin"
)

// FileUpload 本地策略文件上传
func FileUpload(c *gin.Context) {

	// 非本地策略时拒绝上传
	if user, ok := c.Get("user"); ok && user.(*model.User).Policy.Type != "local" {
		c.JSON(200, serializer.Err(serializer.CodePolicyNotAllowed, "当前上传策略无法使用", nil))
		return
	}

	name := c.PostForm("name")
	path := c.PostForm("path")

	// Source
	_, err := c.FormFile("file")
	if err != nil {
		c.JSON(200, serializer.Err(serializer.CodeIOFailed, "无法获取文件数据", err))
		return
	}

	c.JSON(200, serializer.Response{
		Code: 0,
		Msg:  name + path,
	})
}
