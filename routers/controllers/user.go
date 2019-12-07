package controllers

import (
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/service/user"
	"github.com/gin-gonic/gin"
)

// UserLogin 用户登录
func UserLogin(c *gin.Context) {
	var service user.UserLoginService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Login(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}

}

// UserMe 获取当前登录的用户
func UserMe(c *gin.Context) {
	currUser := CurrentUser(c)
	res := serializer.BuildUserResponse(*currUser)
	c.JSON(200, res)
}

// UserStorage 获取用户的存储信息
func UserStorage(c *gin.Context) {
	currUser := CurrentUser(c)
	res := serializer.BuildUserStorageResponse(*currUser)
	c.JSON(200, res)
}
