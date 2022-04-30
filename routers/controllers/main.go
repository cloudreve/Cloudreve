package controllers

import (
	"encoding/json"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ParamErrorMsg 根据Validator返回的错误信息给出错误提示
func ParamErrorMsg(filed string, tag string) string {
	// 未通过验证的表单域与中文对应
	fieldMap := map[string]string{
		"UserName": "Email",
		"Password": "Password",
		"Path":     "Path",
		"SourceID": "Source resource",
		"URL":      "URL",
		"Nick":     "Nickname",
	}
	// 未通过的规则与中文对应
	tagMap := map[string]string{
		"required": "cannot be empty",
		"min":      "too short",
		"max":      "too long",
		"email":    "format error",
	}
	fieldVal, findField := fieldMap[filed]
	if !findField {
		fieldVal = filed
	}
	tagVal, findTag := tagMap[tag]
	if findTag {
		// 返回拼接出来的错误信息
		return fieldVal + " " + tagVal
	}
	return ""
}

// ErrorResponse 返回错误消息
func ErrorResponse(err error) serializer.Response {
	// 处理 Validator 产生的错误
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, e := range ve {
			return serializer.ParamErr(
				ParamErrorMsg(e.Field(), e.Tag()),
				err,
			)
		}
	}

	if _, ok := err.(*json.UnmarshalTypeError); ok {
		return serializer.ParamErr("JSON marshall error", err)
	}

	return serializer.ParamErr("Parameter error", err)
}

// CurrentUser 获取当前用户
func CurrentUser(c *gin.Context) *model.User {
	if user, _ := c.Get("user"); user != nil {
		if u, ok := user.(*model.User); ok {
			return u
		}
	}
	return nil
}
