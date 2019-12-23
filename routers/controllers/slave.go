package controllers

import (
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/gin-gonic/gin"
)

// SlaveUpload 从机文件上传
func SlaveUpload(c *gin.Context) {

	c.JSON(200, serializer.Response{
		Code: 0,
	})
}
