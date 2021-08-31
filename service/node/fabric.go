package node

import (
	"encoding/gob"
	"github.com/cloudreve/Cloudreve/v3/pkg/mq"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/slave"
	"github.com/gin-gonic/gin"
)

type SlaveNotificationService struct {
	Subject string `uri:"subject" binding:"required"`
}

func HandleMasterHeartbeat(req *serializer.NodePingReq) serializer.Response {
	res, err := slave.DefaultController.HandleHeartBeat(req)
	if err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "无法初始化从机控制器", err)
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
		return serializer.ParamErr("无法解析通知消息", err)
	}

	mq.GlobalMQ.Publish(s.Subject, msg)
	return serializer.Response{}
}
