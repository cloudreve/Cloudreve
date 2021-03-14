package callback

import (
	"context"
	"fmt"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/onedrive"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-gonic/gin"
	"strings"
)

// OneDriveOauthService OneDrive 授权回调服务
type OneDriveOauthService struct {
	Code     string `form:"code"`
	Error    string `form:"error"`
	ErrorMsg string `form:"error_description"`
}

// Auth 更新认证信息
func (service *OneDriveOauthService) Auth(c *gin.Context) serializer.Response {
	if service.Error != "" {
		return serializer.ParamErr(service.ErrorMsg, nil)
	}

	policyID, ok := util.GetSession(c, "onedrive_oauth_policy").(uint)
	if !ok {
		return serializer.Err(serializer.CodeNotFound, "授权会话不存在，请重试", nil)
	}

	util.DeleteSession(c, "onedrive_oauth_policy")

	policy, err := model.GetPolicyByID(policyID)
	if err != nil {
		return serializer.Err(serializer.CodeNotFound, "存储策略不存在", nil)
	}

	client, err := onedrive.NewClient(&policy)
	if err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "无法初始化 OneDrive 客户端", err)
	}

	credential, err := client.ObtainToken(c, onedrive.WithCode(service.Code))
	if err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "AccessToken 获取失败", err)
	}

	// 更新存储策略的 RefreshToken
	client.Policy.AccessKey = credential.RefreshToken
	if err := client.Policy.SaveAndClearCache(); err != nil {
		return serializer.DBErr("无法更新 RefreshToken", err)
	}

	cache.Deletes([]string{client.Policy.AccessKey}, "onedrive_")
	if client.Policy.OptionsSerialized.OdDriver != "" && strings.Contains(client.Policy.OptionsSerialized.OdDriver, "http") {
		if err := querySharePointSiteID(c, client.Policy); err != nil {
			return serializer.Err(serializer.CodeInternalSetting, "无法查询 SharePoint 站点 ID", err)
		}
	}

	return serializer.Response{}
}

func querySharePointSiteID(ctx context.Context, policy *model.Policy) error {
	client, err := onedrive.NewClient(policy)
	if err != nil {
		return err
	}

	id, err := client.GetSiteIDByURL(ctx, client.Policy.OptionsSerialized.OdDriver)
	if err != nil {
		return err
	}

	client.Policy.OptionsSerialized.OdDriver = fmt.Sprintf("sites/%s/drive", id)
	if err := client.Policy.SaveAndClearCache(); err != nil {
		return err
	}

	return nil
}
