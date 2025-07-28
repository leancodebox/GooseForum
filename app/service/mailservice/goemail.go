package mailservice

import (
	"fmt"
	"github.com/wneessen/go-mail"
)

func SendV2(to, username, token string) error {
	config := getEmailConfig()
	message := mail.NewMsg()
	if err := message.FromFormat(config.FromName, config.Username); err != nil {
		return fmt.Errorf("failed to set From address: %s", err)
	}
	if err := message.To(to); err != nil {
		return fmt.Errorf("failed to set To address: %s", err)
	}
	message.Subject("activation-email")
	body, err := generateActivationEmailBody(username, token)
	if err != nil {
		return fmt.Errorf("生成邮件内容失败: %v", err)
	}
	message.SetBodyString(mail.TypeTextHTML, body)
	optionList := []mail.Option{
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(config.Username),
		mail.WithPassword(config.Password),
		mail.WithPort(config.Port),
		mail.WithSSL(),
	}
	client, err := mail.NewClient(config.Host,
		optionList...,
	)
	if err != nil {
		return fmt.Errorf("failed to create mail client: %s", err)

	}
	defer client.Close()
	if err = client.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send mail: %s", err)
	}
	return nil
}

// SendTestEmailWithConfig 使用指定配置发送测试邮件
func SendTestEmailWithConfig(config EmailConfig, testEmail string) error {
	// 使用 go-mail 库直接发送测试邮件
	message := mail.NewMsg()
	if err := message.FromFormat(config.FromName, config.FromEmail); err != nil {
		return fmt.Errorf("failed to set From address: %s", err)
	}
	if err := message.To(testEmail); err != nil {
		return fmt.Errorf("failed to set To address: %s", err)
	}
	message.Subject("邮件配置测试")
	message.SetBodyString(mail.TypeTextPlain, "这是一封测试邮件，用于验证邮件配置是否正确。")

	// 创建邮件客户端
	optionList := []mail.Option{
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(config.Username),
		mail.WithPassword(config.Password),
		mail.WithPort(config.Port),
	}

	// 根据加密方式添加选项
	if config.Host != "" {
		optionList = append(optionList, mail.WithSSL())
	}

	client, err := mail.NewClient(config.Host, optionList...)
	if err != nil {
		return fmt.Errorf("failed to create mail client: %s", err)
	}
	defer client.Close()

	if err = client.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send mail: %s", err)
	}

	return nil
}
