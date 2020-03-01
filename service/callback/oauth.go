package callback

import (
	"context"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem/driver/onedrive"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/gin-gonic/gin"
)

// OneDriveOauthService OneDrive 授权回调服务
type OneDriveOauthService struct {
	Code string `form:"code" binding:"required"`
}

// Auth 更新认证信息
func (service *OneDriveOauthService) Auth(c *gin.Context) serializer.Response {
	policyId := util.GetSession(c, "onedrive_oauth_policy").(uint)
	policy, err := model.GetPolicyByID(policyId)
	if err != nil {
		return serializer.Err(serializer.CodeNotFound, "存储策略不存在", nil)
	}

	client, err := onedrive.NewClient(&policy)
	if err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "无法初始化 OneDrive 客户端", err)
	}

	credential, err := client.ObtainToken(context.Background(), onedrive.WithCode(service.Code))
	if err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "AccessToken 获取失败", err)
	}

	// 更新存储策略的 RefreshToken
	if err := client.Policy.UpdateAccessKey(credential.RefreshToken); err != nil {
		return serializer.DBErr("无法更新 RefreshToken", err)
	}

	return serializer.Response{}
}
