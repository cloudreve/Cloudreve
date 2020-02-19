package user

import (
	"crypto/md5"
	"fmt"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/qq"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
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
	Option string `uri:"option" binding:"required,eq=nick|eq=theme|eq=homepage|eq=vip|eq=qq"`
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

// Update 绑定或解绑QQ
func (service *QQBind) Update(c *gin.Context, user *model.User) serializer.Response {
	// 解除绑定
	if user.OpenID != "" {
		if err := user.Update(map[string]interface{}{"open_id": ""}); err != nil {
			return serializer.DBErr("接触绑定失败", err)
		}
		return serializer.Response{}
	}

	// 新建绑定
	res, err := qq.NewLoginRequest()
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, "无法使用QQ登录", err)
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
				return serializer.DBErr("解约失败", err)
			}
		}
	}
	return serializer.Response{}
}

// Update 更改昵称
func (service *ChangerNick) Update(c *gin.Context, user *model.User) serializer.Response {
	if err := user.Update(map[string]interface{}{"nick": service.Nick}); err != nil {
		return serializer.DBErr("无法更新昵称", err)
	}

	return serializer.Response{}
}

// Get 获取用户头像
func (service *AvatarService) Get(c *gin.Context) serializer.Response {
	// 查找目标用户
	uid, _ := c.Get("object_id")
	user, err := model.GetActiveUserByID(uid.(uint))
	if err != nil {
		return serializer.Err(serializer.CodeNotFound, "用户不存在", err)
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
			return serializer.Err(serializer.CodeInternalSetting, "无法解析 Gravatar 服务器地址", err)
		}

		has := md5.Sum([]byte(user.Email))
		avatar, _ := url.Parse(fmt.Sprintf("/avatar/%x?d=mm&s=%s", has, sizes[service.Size]))

		return serializer.Response{
			Code: -301,
			Data: gravatarRoot.ResolveReference(avatar).String(),
		}
	}

	// 本地文件头像
	if user.Avatar == "file" {
		avatarRoot := model.GetSettingByName("avatar_path")
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

	// 取得用户当前策略
	current := user.Policy

	return serializer.BuildPolicySettingRes(available, &current)
}

// Settings 获取用户设定
func (service *SettingService) Settings(c *gin.Context, user *model.User) serializer.Response {
	// 取得存储策略设定
	policy := service.Policy(c, user)

	// 用户组有效期
	var groupExpires int64
	if user.GroupExpires != nil {
		if expires := user.GroupExpires.Unix() - time.Now().Unix(); expires > 0 {
			groupExpires = user.GroupExpires.Unix()
		}
	}

	return serializer.Response{
		Data: map[string]interface{}{
			"policy":        policy.Data.(map[string]interface{}),
			"uid":           user.ID,
			"qq":            user.OpenID != "",
			"homepage":      !user.OptionsSerialized.ProfileOff,
			"two_factor":    user.TwoFactor != "",
			"prefer_theme":  user.OptionsSerialized.PreferredTheme,
			"themes":        model.GetSettingByNames("themes"),
			"group_expires": groupExpires,
		},
	}
}
