package onedrive

// Option 发送请求的额外设置
type Option interface {
	apply(*options)
}

type options struct {
	redirect     string
	code         string
	refreshToken string
}

type optionFunc func(*options)

// WithCode 设置接口Code
func WithCode(t string) Option {
	return optionFunc(func(o *options) {
		o.code = t
	})
}

// WithRefreshToken 设置接口RefreshToken
func WithRefreshToken(t string) Option {
	return optionFunc(func(o *options) {
		o.refreshToken = t
	})
}

func (f optionFunc) apply(o *options) {
	f(o)
}

func newDefaultOption() *options {
	return &options{}
}
