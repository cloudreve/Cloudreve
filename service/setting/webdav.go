package setting

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-gonic/gin"
)

// WebDAVListService WebDAV 列表服务
type WebDAVListService struct {
}

// WebDAVAccountService WebDAV 账号管理服务
type WebDAVAccountService struct {
	ID uint `uri:"id" binding:"required,min=1"`
}

// WebDAVAccountCreateService WebDAV 账号创建服务
type WebDAVAccountCreateService struct {
	Path string `json:"path" binding:"required,min=1,max=65535"`
	Name string `json:"name" binding:"required,min=1,max=255"`
}

// WebDAVMountCreateService WebDAV 挂载创建服务
type WebDAVMountCreateService struct {
	Path   string `json:"path" binding:"required,min=1,max=65535"`
	Policy string `json:"policy" binding:"required,min=1"`
}

// Create 创建WebDAV账户
func (service *WebDAVAccountCreateService) Create(c *gin.Context, user *model.User) serializer.Response {
	account := model.Webdav{
		Name:     service.Name,
		Password: util.RandStringRunes(32),
		UserID:   user.ID,
		Root:     service.Path,
	}

	if _, err := account.Create(); err != nil {
		return serializer.Err(serializer.CodeDBError, "创建失败", err)
	}

	return serializer.Response{
		Data: map[string]interface{}{
			"id":         account.ID,
			"password":   account.Password,
			"created_at": account.CreatedAt,
		},
	}
}

// Delete 删除WebDAV账户
func (service *WebDAVAccountService) Delete(c *gin.Context, user *model.User) serializer.Response {
	model.DeleteWebDAVAccountByID(service.ID, user.ID)
	return serializer.Response{}
}

// Accounts 列出WebDAV账号
func (service *WebDAVListService) Accounts(c *gin.Context, user *model.User) serializer.Response {
	accounts := model.ListWebDAVAccounts(user.ID)

	return serializer.Response{Data: map[string]interface{}{
		"accounts": accounts,
	}}
}
