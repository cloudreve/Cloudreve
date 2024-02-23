package user

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cluster"
	"github.com/cloudreve/Cloudreve/v3/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v3/pkg/qq"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	"github.com/samber/lo"
)

// SettingService 通用设置服务
type SettingService struct {
}

// SettingListService 通用设置列表服务
type SettingListService struct {
	Page int `form:"page" binding:"required,min=1"`
}

// AvatarService 头像服务
type AvatarService struct {
	Size string `uri:"size" binding:"required,eq=l|eq=m|eq=s"`
}

// SettingUpdateService 设定更改服务
type SettingUpdateService struct {
	Option string `uri:"option" binding:"required,eq=nick|eq=theme|eq=homepage|eq=vip|eq=qq|eq=policy|eq=password|eq=2fa|eq=authn"`
}

// OptionsChangeHandler 属性更改接口
type OptionsChangeHandler interface {
	Update(*gin.Context, *model.User) serializer.Response
}

// ChangerNick 昵称更改服务
type ChangerNick struct {
	Nick string `json:"nick" binding:"required,min=1,max=255"`
}

// VIPUnsubscribe 用户组解约服务
type VIPUnsubscribe struct {
}

// QQBind QQ互联服务
type QQBind struct {
}

// PolicyChange 更改存储策略
type PolicyChange struct {
	ID string `json:"id" binding:"required"`
}

// HomePage 更改个人主页开关
type HomePage struct {
	Enabled bool `json:"status"`
}

// PasswordChange 更改密码
type PasswordChange struct {
	Old string `json:"old" binding:"required,min=4,max=64"`
	New string `json:"new" binding:"required,min=4,max=64"`
}

// Enable2FA 开启二步验证
type Enable2FA struct {
	Code string `json:"code" binding:"required"`
}

// DeleteWebAuthn 删除WebAuthn凭证
type DeleteWebAuthn struct {
	ID string `json:"id" binding:"required"`
}

// ThemeChose 主题选择
type ThemeChose struct {
	Theme string `json:"theme" binding:"required,hexcolor|rgb|rgba|hsl"`
}

// Update 更新主题设定
func (service *ThemeChose) Update(c *gin.Context, user *model.User) serializer.Response {
	user.OptionsSerialized.PreferredTheme = service.Theme
	if err := user.UpdateOptions(); err != nil {
		return serializer.DBErr("Failed to update user preferences", err)
	}

	return serializer.Response{}
}

// Update 删除凭证
func (service *DeleteWebAuthn) Update(c *gin.Context, user *model.User) serializer.Response {
	user.RemoveAuthn(service.ID)
	return serializer.Response{}
}

// Update 更改二步验证设定
func (service *Enable2FA) Update(c *gin.Context, user *model.User) serializer.Response {
	if user.TwoFactor == "" {
		// 开启2FA
		secret, ok := util.GetSession(c, "2fa_init").(string)
		if !ok {
			return serializer.Err(serializer.CodeInternalSetting, "You have not initiated 2FA session", nil)
		}

		if !totp.Validate(service.Code, secret) {
			return serializer.ParamErr("Incorrect 2FA code", nil)
		}

		if err := user.Update(map[string]interface{}{"two_factor": secret}); err != nil {
			return serializer.DBErr("Failed to update user preferences", err)
		}

	} else {
		// 关闭2FA
		if !totp.Validate(service.Code, user.TwoFactor) {
			return serializer.ParamErr("Incorrect 2FA code", nil)
		}

		if err := user.Update(map[string]interface{}{"two_factor": ""}); err != nil {
			return serializer.DBErr("Failed to update user preferences", err)
		}
	}

	return serializer.Response{}
}

// Init2FA 初始化二步验证
func (service *SettingService) Init2FA(c *gin.Context, user *model.User) serializer.Response {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Cloudreve",
		AccountName: user.Email,
	})
	if err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "Failed to generate TOTP secret", err)
	}

	util.SetSession(c, map[string]interface{}{"2fa_init": key.Secret()})
	return serializer.Response{Data: key.Secret()}
}

// Update 更改密码
func (service *PasswordChange) Update(c *gin.Context, user *model.User) serializer.Response {
	// 验证老密码
	if ok, _ := user.CheckPassword(service.Old); !ok {
		return serializer.Err(serializer.CodeIncorrectPassword, "", nil)
	}

	// 更改为新密码
	user.SetPassword(service.New)
	if err := user.Update(map[string]interface{}{"password": user.Password}); err != nil {
		return serializer.DBErr("Failed to update password", err)
	}

	return serializer.Response{}
}

// Update 切换个人主页开关
func (service *HomePage) Update(c *gin.Context, user *model.User) serializer.Response {
	user.OptionsSerialized.ProfileOff = !service.Enabled
	if err := user.UpdateOptions(); err != nil {
		return serializer.DBErr("Failed to update user preferences", err)
	}

	return serializer.Response{}
}

// Update 更改用户偏好的存储策略
func (service *PolicyChange) Update(c *gin.Context, user *model.User) serializer.Response {
	// 取得存储策略的ID
	rawID, err := hashid.DecodeHashID(service.ID, hashid.PolicyID)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotExist, "", err)
	}

	// 用户是否可以切换到此存储策略
	if !util.ContainsUint(user.Group.PolicyList, rawID) {
		return serializer.Err(serializer.CodePolicyNotAllowed, "", nil)
	}

	// 查找存储策略
	if _, err := model.GetPolicyByID(rawID); err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, "", nil)
	}

	// 切换存储策略
	user.OptionsSerialized.PreferredPolicy = rawID
	if err := user.UpdateOptions(); err != nil {
		return serializer.DBErr("Failed to update user preferences", err)
	}

	return serializer.Response{}
}

// Update 绑定或解绑QQ
func (service *QQBind) Update(c *gin.Context, user *model.User) serializer.Response {
	// 解除绑定
	if user.OpenID != "" {
		// 只通过QQ登录的用户无法解除绑定
		if strings.HasSuffix(user.Email, "@login.qq.com") {
			return serializer.Err(serializer.CodeNoPermissionErr, "This user cannot be unlinked", nil)
		}

		if err := user.Update(map[string]interface{}{"open_id": ""}); err != nil {
			return serializer.DBErr("Failed to update user open id", err)
		}
		return serializer.Response{
			Data: "",
		}
	}

	// 新建绑定
	res, err := qq.NewLoginRequest()
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, "Failed to start QQ login request", err)
	}

	// 设定QQ登录会话Secret
	util.SetSession(c, map[string]interface{}{"qq_login_secret": res.SecretKey})

	return serializer.Response{
		Data: res.URL,
	}
}

// Update 用户组解约
func (service *VIPUnsubscribe) Update(c *gin.Context, user *model.User) serializer.Response {
	if user.GroupExpires != nil {
		timeNow := time.Now()
		if time.Now().Before(*user.GroupExpires) {
			if err := user.Update(map[string]interface{}{"group_expires": &timeNow}); err != nil {
				return serializer.DBErr("Failed to update user", err)
			}
		}
	}
	return serializer.Response{}
}

// Update 更改昵称
func (service *ChangerNick) Update(c *gin.Context, user *model.User) serializer.Response {
	if err := user.Update(map[string]interface{}{"nick": service.Nick}); err != nil {
		return serializer.DBErr("Failed to update user", err)
	}

	return serializer.Response{}
}

// Get 获取用户头像
func (service *AvatarService) Get(c *gin.Context) serializer.Response {
	// 查找目标用户
	uid, _ := c.Get("object_id")
	user, err := model.GetActiveUserByID(uid.(uint))
	if err != nil {
		return serializer.Err(serializer.CodeUserNotFound, "", err)
	}

	// 未设定头像时，返回404错误
	if user.Avatar == "" {
		c.Status(404)
		return serializer.Response{}
	}

	// 获取头像设置
	sizes := map[string]string{
		"s": model.GetSettingByName("avatar_size_s"),
		"m": model.GetSettingByName("avatar_size_m"),
		"l": model.GetSettingByName("avatar_size_l"),
	}

	// Gravatar 头像重定向
	if user.Avatar == "gravatar" {
		server := model.GetSettingByName("gravatar_server")
		gravatarRoot, err := url.Parse(server)
		if err != nil {
			return serializer.Err(serializer.CodeInternalSetting, "Failed to parse Gravatar server", err)
		}
		email_lowered := strings.ToLower(user.Email)
		has := md5.Sum([]byte(email_lowered))
		avatar, _ := url.Parse(fmt.Sprintf("/avatar/%x?d=mm&s=%s", has, sizes[service.Size]))

		return serializer.Response{
			Code: -301,
			Data: gravatarRoot.ResolveReference(avatar).String(),
		}
	}

	// 本地文件头像
	if user.Avatar == "file" {
		avatarRoot := util.RelativePath(model.GetSettingByName("avatar_path"))
		sizeToInt := map[string]string{
			"s": "0",
			"m": "1",
			"l": "2",
		}

		avatar, err := os.Open(filepath.Join(avatarRoot, fmt.Sprintf("avatar_%d_%s.png", user.ID, sizeToInt[service.Size])))
		if err != nil {
			c.Status(404)
			return serializer.Response{}
		}
		defer avatar.Close()

		http.ServeContent(c.Writer, c.Request, "avatar.png", user.UpdatedAt, avatar)
		return serializer.Response{}
	}

	c.Status(404)
	return serializer.Response{}
}

// ListTasks 列出任务
func (service *SettingListService) ListTasks(c *gin.Context, user *model.User) serializer.Response {
	tasks, total := model.ListTasks(user.ID, service.Page, 10, "updated_at desc")
	return serializer.BuildTaskList(tasks, total)
}

// Policy 获取用户存储策略设置
func (service *SettingService) Policy(c *gin.Context, user *model.User) serializer.Response {
	// 取得用户可用存储策略
	available := make([]model.Policy, 0, len(user.Group.PolicyList))
	for _, id := range user.Group.PolicyList {
		if policy, err := model.GetPolicyByID(id); err == nil {
			available = append(available, policy)
		}
	}

	return serializer.BuildPolicySettingRes(available)
}

// Nodes 获取用户可选节点
func (service *SettingService) Nodes(c *gin.Context, user *model.User) serializer.Response {
	if !user.Group.OptionsSerialized.SelectNode {
		return serializer.Err(serializer.CodeGroupNotAllowed, "", nil)
	}

	availableNodesID := user.Group.OptionsSerialized.AvailableNodes

	// All nodes available
	if len(availableNodesID) == 0 {
		nodes, err := model.GetNodesByStatus(model.NodeActive)
		if err != nil {
			return serializer.DBErr("Failed to list nodes", err)
		}

		availableNodesID = lo.Map[model.Node, uint](nodes, func(node model.Node, index int) uint {
			return node.ID
		})
	}

	// 取得用户可用存储策略
	available := lo.FilterMap[uint, *model.Node](availableNodesID,
		func(id uint, index int) (*model.Node, bool) {
			if node := cluster.Default.GetNodeByID(id); node != nil {
				return node.DBModel(), node.IsActive() && node.IsFeatureEnabled("aria2")
			}

			return nil, false
		})

	return serializer.BuildNodeOptionRes(available)
}

// Settings 获取用户设定
func (service *SettingService) Settings(c *gin.Context, user *model.User) serializer.Response {
	// 用户组有效期
	var groupExpires *time.Time
	if user.GroupExpires != nil {
		if expires := user.GroupExpires.Unix() - time.Now().Unix(); expires > 0 {
			groupExpires = user.GroupExpires
		}
	}

	return serializer.Response{
		Data: map[string]interface{}{
			"uid":           user.ID,
			"qq":            user.OpenID != "",
			"homepage":      !user.OptionsSerialized.ProfileOff,
			"two_factor":    user.TwoFactor != "",
			"prefer_theme":  user.OptionsSerialized.PreferredTheme,
			"themes":        model.GetSettingByName("themes"),
			"group_expires": groupExpires,
			"authn":         serializer.BuildWebAuthnList(user.WebAuthnCredentials()),
		},
	}
}
