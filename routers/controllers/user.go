package controllers

import (
	"Cloudreve/serializer"
	"Cloudreve/service/user"
	"github.com/gin-gonic/gin"
)

// UserLogin 用户登录
func UserLogin(c *gin.Context) {
	var service service.UserLoginService
	if err := c.ShouldBindJSON(&service); err == nil {
		//res := service.Login(c)
		c.JSON(200, serializer.Response{
			Code: 0,
			Msg:  "OK",
		})
	} else {
		c.JSON(200, ErrorResponse(err))
	}

}
