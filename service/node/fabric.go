package node

import (
	"encoding/gob"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cluster"
	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/googledrive"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/onedrive"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/oauth"
	"github.com/cloudreve/Cloudreve/v3/pkg/mq"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/gin-gonic/gin"
)

type SlaveNotificationService struct {
	Subject string `uri:"subject" binding:"required"`
}

type OauthCredentialService struct {
	PolicyID uint `uri:"id" binding:"required"`
}

func HandleMasterHeartbeat(req *serializer.NodePingReq) serializer.Response {
	res, err := cluster.DefaultController.HandleHeartBeat(req)
	if err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "Cannot initialize slave controller", err)
	}

	return serializer.Response{
		Code: 0,
		Data: res,
	}
}

// HandleSlaveNotificationPush 转发从机的消息通知到本机消息队列
func (s *SlaveNotificationService) HandleSlaveNotificationPush(c *gin.Context) serializer.Response {
	var msg mq.Message
	dec := gob.NewDecoder(c.Request.Body)
	if err := dec.Decode(&msg); err != nil {
		return serializer.ParamErr("Cannot parse notification message", err)
	}

	mq.GlobalMQ.Publish(s.Subject, msg)
	return serializer.Response{}
}

// Get 获取主机Oauth策略的AccessToken
func (s *OauthCredentialService) Get(c *gin.Context) serializer.Response {
	policy, err := model.GetPolicyByID(s.PolicyID)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotExist, "", err)
	}

	var client oauth.TokenProvider
	switch policy.Type {
	case "onedrive":
		client, err = onedrive.NewClient(&policy)
		if err != nil {
			return serializer.Err(serializer.CodeInternalSetting, "Cannot initialize OneDrive client", err)
		}
	case "googledrive":
		client, err = googledrive.NewClient(&policy)
		if err != nil {
			return serializer.Err(serializer.CodeInternalSetting, "Cannot initialize Google Drive client", err)
		}
	default:
		return serializer.Err(serializer.CodePolicyNotExist, "", nil)
	}

	if err := client.UpdateCredential(c, conf.SystemConfig.Mode == "slave"); err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "Cannot refresh OneDrive credential", err)
	}

	return serializer.Response{Data: client.AccessToken()}
}
