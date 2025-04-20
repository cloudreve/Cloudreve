package email

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudreve/Cloudreve/v4/inventory"
	"github.com/cloudreve/Cloudreve/v4/pkg/logging"
	"github.com/cloudreve/Cloudreve/v4/pkg/setting"
	"github.com/go-mail/mail"
	"github.com/gofrs/uuid"
)

// SMTPPool SMTP协议发送邮件
type SMTPPool struct {
	// Deprecated
	Config SMTPConfig

	config *setting.SMTP
	ch     chan *message
	chOpen bool
	l      logging.Logger
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
	Keepalive  int    // SMTPPool 连接保留时长
}

type message struct {
	msg    *mail.Message
	cid    string
	userID int
}

// NewSMTPPool initializes a new SMTP based email sending queue.
func NewSMTPPool(config setting.Provider, logger logging.Logger) *SMTPPool {
	client := &SMTPPool{
		config: config.SMTP(context.Background()),
		ch:     make(chan *message, 30),
		chOpen: false,
		l:      logger,
	}

	client.Init()
	return client
}

// NewSMTPClient 新建SMTP发送队列
// Deprecated
func NewSMTPClient(config SMTPConfig) *SMTPPool {
	client := &SMTPPool{
		Config: config,
		ch:     make(chan *message, 30),
		chOpen: false,
	}

	client.Init()

	return client
}

// Send 发送邮件
func (client *SMTPPool) Send(ctx context.Context, to, title, body string) error {
	if !client.chOpen {
		return fmt.Errorf("SMTP pool is closed")
	}

	// 忽略通过QQ登录的邮箱
	if strings.HasSuffix(to, "@login.qq.com") {
		return nil
	}

	m := mail.NewMessage()
	m.SetAddressHeader("From", client.config.From, client.config.FromName)
	m.SetAddressHeader("Reply-To", client.config.ReplyTo, client.config.FromName)
	m.SetHeader("To", to)
	m.SetHeader("Subject", title)
	m.SetHeader("Message-ID", fmt.Sprintf("<%s@%s>", uuid.Must(uuid.NewV4()).String(), "cloudreve"))
	m.SetBody("text/html", body)
	client.ch <- &message{
		msg:    m,
		cid:    logging.CorrelationID(ctx).String(),
		userID: inventory.UserIDFromContext(ctx),
	}
	return nil
}

// Close 关闭发送队列
func (client *SMTPPool) Close() {
	if client.ch != nil {
		close(client.ch)
	}
}

// Init 初始化发送队列
func (client *SMTPPool) Init() {
	go func() {
		client.l.Info("Initializing and starting SMTP email pool...")
		defer func() {
			if err := recover(); err != nil {
				client.chOpen = false
				client.l.Error("Exception while sending email: %s, queue will be reset in 10 seconds.", err)
				time.Sleep(time.Duration(10) * time.Second)
				client.Init()
			}
		}()

		d := mail.NewDialer(client.config.Host, client.config.Port, client.config.User, client.config.Password)
		d.Timeout = time.Duration(client.config.Keepalive+5) * time.Second
		client.chOpen = true
		// 是否启用 SSL
		d.SSL = false
		if client.config.ForceEncryption {
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
					client.l.Info("Email queue closing...")
					client.chOpen = false
					return
				}

				if !open {
					if s, err = d.Dial(); err != nil {
						panic(err)
					}
					open = true
				}

				l := client.l.CopyWithPrefix(fmt.Sprintf("[Cid: %s]", m.cid))
				if err := mail.Send(s, m.msg); err != nil {
					l.Warning("Failed to send email: %s, Cid=%s", err, m.cid)
				} else {
					l.Info("Email sent to %q, title: %q.", m.msg.GetHeader("To"), m.msg.GetHeader("Subject"))
				}
			// 长时间没有新邮件，则关闭SMTP连接
			case <-time.After(time.Duration(client.config.Keepalive) * time.Second):
				if open {
					if err := s.Close(); err != nil {
						client.l.Warning("Failed to close SMTP connection: %s", err)
					}
					open = false
				}
			}
		}
	}()
}
