package serializer

import model "github.com/cloudreve/Cloudreve/v3/models"

// RemoteDeleteRequest 远程策略删除接口请求正文
type RemoteDeleteRequest struct {
	Files []string `json:"files"`
}

// ListRequest 远程策略列文件请求正文
type ListRequest struct {
	Path      string `json:"path"`
	Recursive bool   `json:"recursive"`
}

// NodePingReq 从机节点Ping请求
type NodePingReq struct {
	SiteURL       string      `json:"site_url"`
	SiteID        string      `json:"site_id"`
	IsUpdate      bool        `json:"is_update"`
	CredentialTTL int         `json:"credential_ttl"`
	Node          *model.Node `json:"node"`
}

// NodePingResp 从机节点Ping响应
type NodePingResp struct {
}

// SlaveAria2Call 从机有关Aria2的请求正文
type SlaveAria2Call struct {
	Task         *model.Download        `json:"task"`
	GroupOptions map[string]interface{} `json:"group_options"`
	Files        []int                  `json:"files"`
}
