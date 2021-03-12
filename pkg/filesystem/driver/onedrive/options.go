package onedrive

import "time"

// Option 发送请求的额外设置
type Option interface {
	apply(*options)
}

type options struct {
	redirect          string
	code              string
	refreshToken      string
	conflictBehavior  string
	expires           time.Time
	useDriverResource bool
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

// WithConflictBehavior 设置文件重名后的处理方式
func WithConflictBehavior(t string) Option {
	return optionFunc(func(o *options) {
		o.conflictBehavior = t
	})
}

// WithConflictBehavior 设置文件重名后的处理方式
func WithDriverResource(t bool) Option {
	return optionFunc(func(o *options) {
		o.useDriverResource = t
	})
}

func (f optionFunc) apply(o *options) {
	f(o)
}

func newDefaultOption() *options {
	return &options{
		conflictBehavior:  "fail",
		useDriverResource: true,
		expires:           time.Now().UTC().Add(time.Duration(1) * time.Hour),
	}
}
