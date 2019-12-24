package serializer

// UploadPolicy slave模式下传递的上传策略
type UploadPolicy struct {
	SavePath         string   `json:"save_path"`
	MaxSize          uint64   `json:"save_path"`
	AllowedExtension []string `json:"allowed_extension"`
	CallbackURL      string   `json:"callback_url"`
	CallbackKey      string   `json:"callback_key"`
}
