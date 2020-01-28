package controllers

import (
	"encoding/json"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/authn"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/HFO4/cloudreve/service/user"
	"github.com/duo-labs/webauthn/webauthn"
	"github.com/gin-gonic/gin"
)

// StartLoginAuthn 开始注册WebAuthn登录
func StartLoginAuthn(c *gin.Context) {
	userName := c.Param("username")
	expectedUser, err := model.GetUserByEmail(userName)
	if err != nil {
		c.JSON(200, serializer.Err(401, "用户邮箱或密码错误", err))
		return
	}

	options, sessionData, err := authn.AuthnInstance.BeginLogin(expectedUser)
	if err != nil {
		c.JSON(200, ErrorResponse(err))
		return
	}

	val, err := json.Marshal(sessionData)
	if err != nil {
		c.JSON(200, ErrorResponse(err))
		return
	}

	util.SetSession(c, map[string]interface{}{
		"registration-session": val,
	})
	c.JSON(200, serializer.Response{Code: 0, Data: options})
}

// FinishLoginAuthn 完成注册WebAuthn登录
func FinishLoginAuthn(c *gin.Context) {
	userName := c.Param("username")
	expectedUser, err := model.GetUserByEmail(userName)
	if err != nil {
		c.JSON(200, serializer.Err(401, "用户邮箱或密码错误", err))
		return
	}

	sessionDataJSON := util.GetSession(c, "registration-session").([]byte)

	var sessionData webauthn.SessionData
	err = json.Unmarshal(sessionDataJSON, &sessionData)

	_, err = authn.AuthnInstance.FinishLogin(expectedUser, sessionData, c.Request)

	if err != nil {
		c.JSON(200, serializer.Err(401, "用户邮箱或密码错误", err))
		return
	}

	util.SetSession(c, map[string]interface{}{
		"user_id": expectedUser.ID,
	})
	c.JSON(200, serializer.BuildUserResponse(expectedUser))
}

// StartRegAuthn 开始注册WebAuthn信息
func StartRegAuthn(c *gin.Context) {
	currUser := CurrentUser(c)
	options, sessionData, err := authn.AuthnInstance.BeginRegistration(currUser)
	if err != nil {
		c.JSON(200, ErrorResponse(err))
		return
	}

	val, err := json.Marshal(sessionData)
	if err != nil {
		c.JSON(200, ErrorResponse(err))
		return
	}

	util.SetSession(c, map[string]interface{}{
		"registration-session": val,
	})
	c.JSON(200, serializer.Response{Code: 0, Data: options})
}

// FinishRegAuthn 完成注册WebAuthn信息
func FinishRegAuthn(c *gin.Context) {
	currUser := CurrentUser(c)
	sessionDataJSON := util.GetSession(c, "registration-session").([]byte)

	var sessionData webauthn.SessionData
	err := json.Unmarshal(sessionDataJSON, &sessionData)

	credential, err := authn.AuthnInstance.FinishRegistration(currUser, sessionData, c.Request)

	currUser.RegisterAuthn(credential)
	if err != nil {
		c.JSON(200, ErrorResponse(err))
		return
	}
	c.JSON(200, serializer.Response{Code: 0})
}

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

// UserSignOut 用户退出登录
func UserSignOut(c *gin.Context) {
	util.DeleteSession(c, "user_id")
	c.JSON(200, serializer.Response{})
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
