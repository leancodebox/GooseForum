package mailservice

import (
	"errors"
	"fmt"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/wneessen/go-mail"
	"time"
)

func buildClientByConfig(config pageConfig.MailSettingsConfig) (*mail.Client, error) {
	optionList := []mail.Option{
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(config.SmtpUsername),
		mail.WithPassword(config.SmtpPassword),
		mail.WithPort(config.SmtpPort),
	}

	// 根据加密方式添加选项
	if config.UseSSL {
		optionList = append(optionList, mail.WithSSL())
	}

	return mail.NewClient(config.SmtpHost, optionList...)
}

func SendActivationEmail(to, username, token string) error {
	config := hotdataserve.GetMailSettingsConfigCache()
	if !config.EnableMail {
		return errors.New("Mail settings config is disabled")
	}
	message := mail.NewMsg()
	if err := message.To(to); err != nil {
		return fmt.Errorf("failed to set To address: %s", err)
	}
	message.Subject("activation-email")
	if err := message.FromFormat(config.FromName, config.FromEmail); err != nil {
		return fmt.Errorf("failed to set From address: %s", err)
	}
	body, err := generateActivationEmailBody(username, token)
	if err != nil {
		return fmt.Errorf("生成邮件内容失败: %v", err)
	}
	message.SetBodyString(mail.TypeTextHTML, body)

	client, err := buildClientByConfig(config)
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
func SendTestEmailWithConfig(config pageConfig.MailSettingsConfig, testEmail string) error {
	// 使用 go-mail 库直接发送测试邮件
	message := mail.NewMsg()
	if err := message.FromFormat(config.FromName, config.FromEmail); err != nil {
		return fmt.Errorf("failed to set From address: %s", err)
	}
	if err := message.To(testEmail); err != nil {
		return fmt.Errorf("failed to set To address: %s", err)
	}
	message.Subject("邮件配置测试")
	message.SetBodyString(mail.TypeTextPlain,
		fmt.Sprintf("[%v]这是一封测试邮件,用于验证邮件配置是否正确。", time.Now().Format(time.DateTime)))

	client, err := buildClientByConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create mail client: %s", err)
	}
	defer client.Close()

	if err = client.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send mail: %s", err)
	}

	return nil
}
