package email

import (
	"errors"
	"strings"
)

// Driver 邮件发送驱动
type Driver interface {
	// Close 关闭驱动
	Close()
	// Send 发送邮件
	Send(to, title, body string) error
}

var (
	// ErrChanNotOpen 邮件队列未开启
	ErrChanNotOpen = errors.New("email queue is not started")
	// ErrNoActiveDriver 无可用邮件发送服务
	ErrNoActiveDriver = errors.New("no avaliable email provider")
)

// Send 发送邮件
func Send(to, title, body string) error {
	// 忽略通过QQ登录的邮箱
	if strings.HasSuffix(to, "@login.qq.com") {
		return nil
	}

	Lock.RLock()
	defer Lock.RUnlock()

	if Client == nil {
		return ErrNoActiveDriver
	}

	return Client.Send(to, title, body)
}
