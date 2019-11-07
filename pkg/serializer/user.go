package serializer

// UserOption 用户个性化配置字段
type UserOption struct {
	ProfileOn int    `json:"profile_on"`
	WebDAVKey string `json:"webdav_key"`
}
