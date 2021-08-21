package node

import (
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/slave"
)

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
