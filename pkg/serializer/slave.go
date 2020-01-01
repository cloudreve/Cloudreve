package serializer

// RemoteDeleteRequest 远程策略删除接口请求正文
type RemoteDeleteRequest struct {
	Files []string `json:"files"`
}
