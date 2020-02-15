package email

import "errors"

// Driver 邮件发送驱动
type Driver interface {
	// Close 关闭驱动
	Close()
	// Send 发送邮件
	Send(to, title, body string) error
}

var (
	// ErrChanNotOpen 邮件队列未开启
	ErrChanNotOpen = errors.New("邮件队列未开启")
	// ErrNoActiveDriver 无可用邮件发送服务
	ErrNoActiveDriver = errors.New("无可用邮件发送服务")
)

// Send 发送邮件
func Send(to, title, body string) error {
	if Client == nil {
		return ErrNoActiveDriver
	}

	return Client.Send(to, title, body)
}
