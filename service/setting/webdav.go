package setting

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/hashid"
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

// WebDAVAccountUpdateReadonlyService WebDAV 修改只读性服务
type WebDAVAccountUpdateReadonlyService struct {
	ID       uint `json:"id" binding:"required,min=1"`
	Readonly bool `json:"readonly"`
}

// WebDAVMountCreateService WebDAV 挂载创建服务
type WebDAVMountCreateService struct {
	Path   string `json:"path" binding:"required,min=1,max=65535"`
	Policy string `json:"policy" binding:"required,min=1"`
}

// Create 创建目录挂载
func (service *WebDAVMountCreateService) Create(c *gin.Context, user *model.User) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystem(user)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}
	defer fs.Recycle()

	// 检索要挂载的目录
	exist, folder := fs.IsPathExist(service.Path)
	if !exist {
		return serializer.Err(serializer.CodeParentNotExist, "", err)
	}

	// 检索要挂载的存储策略
	policyID, err := hashid.DecodeHashID(service.Policy, hashid.PolicyID)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotExist, "", err)
	}

	// 检查存储策略是否可用
	if policy, err := model.GetPolicyByID(policyID); err != nil || !util.ContainsUint(user.Group.PolicyList, policy.ID) {
		return serializer.Err(serializer.CodePolicyNotAllowed, "", err)
	}

	// 挂载
	if err := folder.Mount(policyID); err != nil {
		return serializer.Err(serializer.CodeDBError, "Failed to update folder record", err)
	}

	return serializer.Response{
		Data: map[string]interface{}{
			"id": hashid.HashID(folder.ID, hashid.FolderID),
		},
	}
}

// Unmount 取消目录挂载
func (service *WebDAVListService) Unmount(c *gin.Context, user *model.User) serializer.Response {
	folderID, _ := c.Get("object_id")
	folder, err := model.GetFoldersByIDs([]uint{folderID.(uint)}, user.ID)
	if err != nil || len(folder) == 0 {
		return serializer.Err(serializer.CodeParentNotExist, "", err)
	}

	if err := folder[0].Mount(0); err != nil {
		return serializer.DBErr("Failed to update folder record", err)
	}

	return serializer.Response{}
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
		return serializer.DBErr("Failed to create account record", err)
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

// Update 修改WebDAV账户的只读性
func (service *WebDAVAccountUpdateReadonlyService) Update(c *gin.Context, user *model.User) serializer.Response {
	model.UpdateWebDAVAccountReadonlyByID(service.ID, user.ID, service.Readonly)
	return serializer.Response{Data: map[string]bool{
		"readonly": service.Readonly,
	}}
}

// Accounts 列出WebDAV账号
func (service *WebDAVListService) Accounts(c *gin.Context, user *model.User) serializer.Response {
	accounts := model.ListWebDAVAccounts(user.ID)

	// 查找挂载了存储策略的目录
	folders := model.GetMountedFolders(user.ID)

	return serializer.Response{Data: map[string]interface{}{
		"accounts": accounts,
		"folders":  serializer.BuildMountedFolderRes(folders, user.Group.PolicyList),
	}}
}
