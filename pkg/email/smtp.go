package email

import (
	"time"

	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/go-mail/mail"
)

// SMTP SMTP协议发送邮件
type SMTP struct {
	Config SMTPConfig
	ch     chan *mail.Message
	chOpen bool
}

// SMTPConfig SMTP发送配置
type SMTPConfig struct {
	Name       string // 发送者名
	Address    string // 发送者地址
	ReplyTo    string // 回复地址
	Host       string // 服务器主机名
	Port       int    // 服务器端口
	User       string // 用户名
	Password   string // 密码
	Encryption bool   // 是否启用加密
	Keepalive  int    // SMTP 连接保留时长
}

// NewSMTPClient 新建SMTP发送队列
func NewSMTPClient(config SMTPConfig) *SMTP {
	client := &SMTP{
		Config: config,
		ch:     make(chan *mail.Message, 30),
		chOpen: false,
	}

	client.Init()

	return client
}

// Send 发送邮件
func (client *SMTP) Send(to, title, body string) error {
	if !client.chOpen {
		return ErrChanNotOpen
	}
	m := mail.NewMessage()
	m.SetAddressHeader("From", client.Config.Address, client.Config.Name)
	m.SetAddressHeader("Reply-To", client.Config.ReplyTo, client.Config.Name)
	m.SetHeader("To", to)
	m.SetHeader("Subject", title)
	m.SetBody("text/html", body)
	client.ch <- m
	return nil
}

// Close 关闭发送队列
func (client *SMTP) Close() {
	if client.ch != nil {
		close(client.ch)
	}
}

// Init 初始化发送队列
func (client *SMTP) Init() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				client.chOpen = false
				util.Log().Error("邮件发送队列出现异常, %s ,10 秒后重置", err)
				time.Sleep(time.Duration(10) * time.Second)
				client.Init()
			}
		}()

		d := mail.NewDialer(client.Config.Host, client.Config.Port, client.Config.User, client.Config.Password)
		d.Timeout = time.Duration(client.Config.Keepalive+5) * time.Second
		client.chOpen = true
		// 是否启用 SSL
		d.SSL = false
		if client.Config.Encryption {
			d.SSL = true
		}
		d.StartTLSPolicy = mail.OpportunisticStartTLS

		var s mail.SendCloser
		var err error
		open := false
		for {
			select {
			case m, ok := <-client.ch:
				if !ok {
					util.Log().Debug("邮件队列关闭")
					client.chOpen = false
					return
				}
				if !open {
					if s, err = d.Dial(); err != nil {
						panic(err)
					}
					open = true
				}
				if err := mail.Send(s, m); err != nil {
					util.Log().Warning("邮件发送失败, %s", err)
				} else {
					util.Log().Debug("邮件已发送")
				}
			// 长时间没有新邮件，则关闭SMTP连接
			case <-time.After(time.Duration(client.Config.Keepalive) * time.Second):
				if open {
					if err := s.Close(); err != nil {
						util.Log().Warning("无法关闭 SMTP 连接 %s", err)
					}
					open = false
				}
			}
		}
	}()
}
