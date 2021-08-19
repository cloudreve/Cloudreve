package serializer

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
}

// NodePingResp 从机节点Ping响应
type NodePingResp struct {
}
