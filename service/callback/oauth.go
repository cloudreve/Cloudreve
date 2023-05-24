package callback

import (
	"context"
	"fmt"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/googledrive"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/onedrive"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"strings"
)

// OauthService OAuth 存储策略授权回调服务
type OauthService struct {
	Code     string `form:"code"`
	Error    string `form:"error"`
	ErrorMsg string `form:"error_description"`
	Scope    string `form:"scope"`
}

// GDriveAuth Google Drive 更新认证信息
func (service *OauthService) GDriveAuth(c *gin.Context) serializer.Response {
	if service.Error != "" {
		return serializer.ParamErr(service.Error, nil)
	}

	// validate required scope
	if missing, found := lo.Find[string](googledrive.RequiredScope, func(item string) bool {
		return !strings.Contains(service.Scope, item)
	}); found {
		return serializer.ParamErr(fmt.Sprintf("Missing required scope: %s", missing), nil)
	}

	policyID, ok := util.GetSession(c, "googledrive_oauth_policy").(uint)
	if !ok {
		return serializer.Err(serializer.CodeNotFound, "", nil)
	}

	util.DeleteSession(c, "googledrive_oauth_policy")

	policy, err := model.GetPolicyByID(policyID)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotExist, "", nil)
	}

	client, err := googledrive.NewClient(&policy)
	if err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "Failed to initialize Google Drive client", err)
	}

	credential, err := client.ObtainToken(c, service.Code, "")
	if err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "Failed to fetch AccessToken", err)
	}

	// 更新存储策略的 RefreshToken
	client.Policy.AccessKey = credential.RefreshToken
	if err := client.Policy.SaveAndClearCache(); err != nil {
		return serializer.DBErr("Failed to update RefreshToken", err)
	}

	cache.Deletes([]string{client.Policy.AccessKey}, googledrive.TokenCachePrefix)
	return serializer.Response{}
}

// OdAuth OneDrive 更新认证信息
func (service *OauthService) OdAuth(c *gin.Context) serializer.Response {
	if service.Error != "" {
		return serializer.ParamErr(service.ErrorMsg, nil)
	}

	policyID, ok := util.GetSession(c, "onedrive_oauth_policy").(uint)
	if !ok {
		return serializer.Err(serializer.CodeNotFound, "", nil)
	}

	util.DeleteSession(c, "onedrive_oauth_policy")

	policy, err := model.GetPolicyByID(policyID)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotExist, "", nil)
	}

	client, err := onedrive.NewClient(&policy)
	if err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "Failed to initialize OneDrive client", err)
	}

	credential, err := client.ObtainToken(c, onedrive.WithCode(service.Code))
	if err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "Failed to fetch AccessToken", err)
	}

	// 更新存储策略的 RefreshToken
	client.Policy.AccessKey = credential.RefreshToken
	if err := client.Policy.SaveAndClearCache(); err != nil {
		return serializer.DBErr("Failed to update RefreshToken", err)
	}

	cache.Deletes([]string{client.Policy.AccessKey}, "onedrive_")
	if client.Policy.OptionsSerialized.OdDriver != "" && strings.Contains(client.Policy.OptionsSerialized.OdDriver, "http") {
		if err := querySharePointSiteID(c, client.Policy); err != nil {
			return serializer.Err(serializer.CodeInternalSetting, "Failed to query SharePoint site ID", err)
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
