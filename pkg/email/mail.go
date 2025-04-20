package email

import (
	"context"
	"errors"
)

// Driver 邮件发送驱动
type Driver interface {
	// Close 关闭驱动
	Close()
	// Send 发送邮件
	Send(ctx context.Context, to, title, body string) error
}

var (
	// ErrChanNotOpen 邮件队列未开启
	ErrChanNotOpen = errors.New("email queue is not started")
	// ErrNoActiveDriver 无可用邮件发送服务
	ErrNoActiveDriver = errors.New("no avaliable email provider")
)
