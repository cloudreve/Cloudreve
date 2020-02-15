package email

import model "github.com/HFO4/cloudreve/models"

// Client 默认的邮件发送客户端
var Client Driver

// Init 初始化
func Init() {
	if Client != nil {
		Client.Close()
	}

	// 读取SMTP设置
	options := model.GetSettingByNames(
		"fromName",
		"fromAdress",
		"smtpHost",
		"replyTo",
		"smtpUser",
		"smtpPass",
	)
	port := model.GetIntSetting("smtpPort", 25)
	keepAlive := model.GetIntSetting("mail_keepalive", 30)

	client := NewSMTPClient(SMTPConfig{
		Name:      options["fromName"],
		Address:   options["fromAdress"],
		ReplyTo:   options["replyTo"],
		Host:      options["smtpHost"],
		Port:      port,
		User:      options["smtpUser"],
		Password:  options["smtpPass"],
		Keepalive: keepAlive,
	})

	Client = client
}
