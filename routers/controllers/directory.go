package controllers

import (
	"errors"
	"github.com/cloudreve/Cloudreve/v4/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v4/service/explorer"
	"github.com/gin-gonic/gin"
)

// ListDirectory 列出目录下内容
func ListDirectory(c *gin.Context) {
	service := ParametersFromContext[*explorer.ListFileService](c, explorer.ListFileParameterCtx{})
	resp, err := service.List(c)
	if err != nil {
		if errors.Is(err, explorer.ErrSSETakeOver) {
			return
		}

		c.JSON(200, serializer.Err(c, err))
		c.Abort()
		return
	}

	c.JSON(200, serializer.Response{
		Data: resp,
	})
}
