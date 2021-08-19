package node

import "github.com/cloudreve/Cloudreve/v3/pkg/serializer"

func HandleMasterHeartbeat(req *serializer.NodePingReq) serializer.Response {
	return serializer.Response{
		Code: 0,
		Data: serializer.NodePingResp{},
	}
}
