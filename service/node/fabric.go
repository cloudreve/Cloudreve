package node

import (
	"encoding/gob"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cluster"
	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/onedrive"
	"github.com/cloudreve/Cloudreve/v3/pkg/mq"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/gin-gonic/gin"
)

type SlaveNotificationService struct {
	Subject string `uri:"subject" binding:"required"`
}

type OneDriveCredentialService struct {
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

// Get 获取主机OneDrive策略的AccessToken
func (s *OneDriveCredentialService) Get(c *gin.Context) serializer.Response {
	policy, err := model.GetPolicyByID(s.PolicyID)
	if err != nil {
		return serializer.Err(serializer.CodeNotFound, "Cannot found storage policy", err)
	}

	client, err := onedrive.NewClient(&policy)
	if err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "Cannot initialize OneDrive client", err)
	}

	if err := client.UpdateCredential(c, conf.SystemConfig.Mode == "slave"); err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "Cannot refresh OneDrive credential", err)
	}

	return serializer.Response{Data: client.Credential.AccessToken}
}
