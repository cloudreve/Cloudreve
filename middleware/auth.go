package middleware

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/http"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/onedrive"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/oss"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/upyun"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/qiniu/api.v7/v7/auth/qbox"
)

// SignRequired 验证请求签名
func SignRequired(authInstance auth.Auth) gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		switch c.Request.Method {
		case "PUT", "POST", "PATCH":
			err = auth.CheckRequest(authInstance, c.Request)
		default:
			err = auth.CheckURI(authInstance, c.Request.URL)
		}

		if err != nil {
			c.JSON(200, serializer.Err(serializer.CodeCredentialInvalid, err.Error(), err))
			c.Abort()
			return
		}

		c.Next()
	}
}

// CurrentUser 获取登录用户
func CurrentUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		uid := session.Get("user_id")
		if uid != nil {
			user, err := model.GetActiveUserByID(uid)
			if err == nil {
				c.Set("user", &user)
			}
		}
		c.Next()
	}
}

// AuthRequired 需要登录
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if user, _ := c.Get("user"); user != nil {
			if _, ok := user.(*model.User); ok {
				c.Next()
				return
			}
		}

		c.JSON(200, serializer.CheckLogin())
		c.Abort()
	}
}

// WebDAVAuth 验证WebDAV登录及权限
func WebDAVAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// OPTIONS 请求不需要鉴权，否则Windows10下无法保存文档
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		username, password, ok := c.Request.BasicAuth()
		if !ok {
			c.Writer.Header()["WWW-Authenticate"] = []string{`Basic realm="cloudreve"`}
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}

		expectedUser, err := model.GetActiveUserByEmail(username)
		if err != nil {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}

		// 密码正确？
		webdav, err := model.GetWebdavByPassword(password, expectedUser.ID)
		if err != nil {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}

		// 用户组已启用WebDAV？
		if !expectedUser.Group.WebDAVEnabled {
			c.Status(http.StatusForbidden)
			c.Abort()
			return
		}

		c.Set("user", &expectedUser)
		c.Set("webdav", webdav)
		c.Next()
	}
}

// uploadCallbackCheck 对上传回调请求的 callback key 进行验证，如果成功则返回上传用户
func uploadCallbackCheck(c *gin.Context) (serializer.Response, *model.User) {
	// 验证 Callback Key
	callbackKey := c.Param("key")
	if callbackKey == "" {
		return serializer.ParamErr("Callback Key 不能为空", nil), nil
	}
	callbackSessionRaw, exist := cache.Get("callback_" + callbackKey)
	if !exist {
		return serializer.ParamErr("回调会话不存在或已过期", nil), nil
	}
	callbackSession := callbackSessionRaw.(serializer.UploadSession)
	c.Set("callbackSession", &callbackSession)

	// 清理回调会话
	_ = cache.Deletes([]string{callbackKey}, "callback_")

	// 查找用户
	user, err := model.GetActiveUserByID(callbackSession.UID)
	if err != nil {
		return serializer.Err(serializer.CodeCheckLogin, "找不到用户", err), nil
	}
	c.Set("user", &user)

	return serializer.Response{}, &user
}

// RemoteCallbackAuth 远程回调签名验证
func RemoteCallbackAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证key并查找用户
		resp, user := uploadCallbackCheck(c)
		if resp.Code != 0 {
			c.JSON(200, resp)
			c.Abort()
			return
		}

		// 验证签名
		authInstance := auth.HMACAuth{SecretKey: []byte(user.Policy.SecretKey)}
		if err := auth.CheckRequest(authInstance, c.Request); err != nil {
			c.JSON(200, serializer.Err(serializer.CodeCheckLogin, err.Error(), err))
			c.Abort()
			return
		}

		c.Next()

	}
}

// QiniuCallbackAuth 七牛回调签名验证
func QiniuCallbackAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证key并查找用户
		resp, user := uploadCallbackCheck(c)
		if resp.Code != 0 {
			c.JSON(401, serializer.GeneralUploadCallbackFailed{Error: resp.Msg})
			c.Abort()
			return
		}

		// 验证回调是否来自qiniu
		mac := qbox.NewMac(user.Policy.AccessKey, user.Policy.SecretKey)
		ok, err := mac.VerifyCallback(c.Request)
		if err != nil {
			util.Log().Debug("无法验证回调请求，%s", err)
			c.JSON(401, serializer.GeneralUploadCallbackFailed{Error: "无法验证回调请求"})
			c.Abort()
			return
		}
		if !ok {
			c.JSON(401, serializer.GeneralUploadCallbackFailed{Error: "回调签名无效"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OSSCallbackAuth 阿里云OSS回调签名验证
func OSSCallbackAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证key并查找用户
		resp, _ := uploadCallbackCheck(c)
		if resp.Code != 0 {
			c.JSON(401, serializer.GeneralUploadCallbackFailed{Error: resp.Msg})
			c.Abort()
			return
		}

		err := oss.VerifyCallbackSignature(c.Request)
		if err != nil {
			util.Log().Debug("回调签名验证失败，%s", err)
			c.JSON(401, serializer.GeneralUploadCallbackFailed{Error: "回调签名验证失败"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// UpyunCallbackAuth 又拍云回调签名验证
func UpyunCallbackAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证key并查找用户
		resp, user := uploadCallbackCheck(c)
		if resp.Code != 0 {
			c.JSON(401, serializer.GeneralUploadCallbackFailed{Error: resp.Msg})
			c.Abort()
			return
		}

		// 获取请求正文
		body, err := ioutil.ReadAll(c.Request.Body)
		c.Request.Body.Close()
		if err != nil {
			c.JSON(401, serializer.GeneralUploadCallbackFailed{Error: err.Error()})
			c.Abort()
			return
		}

		c.Request.Body = ioutil.NopCloser(bytes.NewReader(body))

		// 准备验证Upyun回调签名
		handler := upyun.Driver{Policy: &user.Policy}
		contentMD5 := c.Request.Header.Get("Content-Md5")
		date := c.Request.Header.Get("Date")
		actualSignature := c.Request.Header.Get("Authorization")

		// 计算正文MD5
		actualContentMD5 := fmt.Sprintf("%x", md5.Sum(body))
		if actualContentMD5 != contentMD5 {
			c.JSON(401, serializer.GeneralUploadCallbackFailed{Error: "MD5不一致"})
			c.Abort()
			return
		}

		// 计算理论签名
		signature := handler.Sign(context.Background(), []string{
			"POST",
			c.Request.URL.Path,
			date,
			contentMD5,
		})

		// 对比签名
		if signature != actualSignature {
			c.JSON(401, serializer.GeneralUploadCallbackFailed{Error: "鉴权失败"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OneDriveCallbackAuth OneDrive回调签名验证
// TODO 解耦
func OneDriveCallbackAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证key并查找用户
		resp, _ := uploadCallbackCheck(c)
		if resp.Code != 0 {
			c.JSON(401, serializer.GeneralUploadCallbackFailed{Error: resp.Msg})
			c.Abort()
			return
		}

		// 发送回调结束信号
		onedrive.FinishCallback(c.Param("key"))

		c.Next()
	}
}

// COSCallbackAuth 腾讯云COS回调签名验证
// TODO 解耦 测试
func COSCallbackAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证key并查找用户
		resp, _ := uploadCallbackCheck(c)
		if resp.Code != 0 {
			c.JSON(401, serializer.GeneralUploadCallbackFailed{Error: resp.Msg})
			c.Abort()
			return
		}

		c.Next()
	}
}

// S3CallbackAuth Amazon S3回调签名验证
func S3CallbackAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证key并查找用户
		resp, _ := uploadCallbackCheck(c)
		if resp.Code != 0 {
			c.JSON(401, serializer.GeneralUploadCallbackFailed{Error: resp.Msg})
			c.Abort()
			return
		}

		c.Next()
	}
}

// IsAdmin 必须为管理员用户组
func IsAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := c.Get("user")
		if user.(*model.User).Group.ID != 1 && user.(*model.User).ID != 1 {
			c.JSON(200, serializer.Err(serializer.CodeAdminRequired, "您不是管理组成员", nil))
			c.Abort()
			return
		}

		c.Next()
	}
}
