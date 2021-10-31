package serializer

import (
	"crypto/sha1"
	"encoding/gob"
	"fmt"
	model "github.com/cloudreve/Cloudreve/v3/models"
)

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

// SlaveTransferReq 从机中转任务创建请求
type SlaveTransferReq struct {
	Src    string        `json:"src"`
	Dst    string        `json:"dst"`
	Policy *model.Policy `json:"policy"`
}

// Hash 返回创建请求的唯一标识，保持创建请求幂等
func (s *SlaveTransferReq) Hash(id string) string {
	h := sha1.New()
	h.Write([]byte(fmt.Sprintf("transfer-%s-%s-%s-%d", id, s.Src, s.Dst, s.Policy.ID)))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

const (
	SlaveTransferSuccess = "success"
	SlaveTransferFailed  = "failed"
)

type SlaveTransferResult struct {
	Error string
}

func init() {
	gob.Register(SlaveTransferResult{})
}
