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

// WebDAVAccountUpdateService WebDAV 修改只读性和是否使用代理服务
type WebDAVAccountUpdateService struct {
	ID       uint  `json:"id" binding:"required,min=1"`
	Readonly *bool `json:"readonly" binding:"required_without=UseProxy"`
	UseProxy *bool `json:"use_proxy" binding:"required_without=Readonly"`
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

// Update 修改WebDAV账户只读性和是否使用代理服务
func (service *WebDAVAccountUpdateService) Update(c *gin.Context, user *model.User) serializer.Response {
	var updates = make(map[string]interface{})
	if service.Readonly != nil {
		updates["readonly"] = *service.Readonly
	}
	if service.UseProxy != nil {
		updates["use_proxy"] = *service.UseProxy
	}
	model.UpdateWebDAVAccountByID(service.ID, user.ID, updates)
	return serializer.Response{Data: updates}
}

// Accounts 列出WebDAV账号
func (service *WebDAVListService) Accounts(c *gin.Context, user *model.User) serializer.Response {
	accounts := model.ListWebDAVAccounts(user.ID)

	return serializer.Response{Data: map[string]interface{}{
		"accounts": accounts,
	}}
}
