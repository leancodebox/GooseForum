package mailservice

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/models/defaultconfig"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/wneessen/go-mail"
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

func normalizeSender(config pageConfig.MailSettingsConfig) (string, string) {
	defaultConfig := defaultconfig.GetDefaultEmailSettingsConfig()
	fromName := strings.TrimSpace(config.FromName)
	if fromName == "" {
		fromName = defaultConfig.FromName
	}

	fromEmail := strings.TrimSpace(config.FromEmail)
	if fromEmail == "" {
		fromEmail = strings.TrimSpace(config.SmtpUsername)
	}
	if fromEmail == "" {
		fromEmail = strings.TrimSpace(defaultConfig.FromEmail)
	}
	if fromEmail == "" {
		fromEmail = "noreply@localhost"
	}

	return fromName, fromEmail
}

func setMessageFrom(message *mail.Msg, config pageConfig.MailSettingsConfig) error {
	fromName, fromEmail := normalizeSender(config)
	if err := message.FromFormat(fromName, fromEmail); err != nil {
		return fmt.Errorf("failed to set From address: %s", err)
	}
	return nil
}

func SendActivationEmail(to, username, token string) error {
	config := hotdataserve.GetMailSettingsConfigCache()
	fromName, fromEmail := normalizeSender(config)
	slog.Debug("准备发送激活邮件", "to", to, "username", username, "enableMail", config.EnableMail, "smtpHost", config.SmtpHost, "smtpPort", config.SmtpPort, "fromName", fromName, "fromEmail", fromEmail)
	if !config.EnableMail {
		return errors.New("mail settings config is disabled")
	}
	message := mail.NewMsg()
	if err := message.To(to); err != nil {
		return fmt.Errorf("failed to set To address: %s", err)
	}
	message.Subject("activation-email")
	if err := setMessageFrom(message, config); err != nil {
		return err
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
		slog.Debug("激活邮件 SMTP 发送失败", "to", to, "username", username, "err", err)
		return fmt.Errorf("failed to send mail: %s", err)
	}
	slog.Debug("激活邮件 SMTP 发送成功", "to", to, "username", username)
	return nil
}

// SendPasswordResetEmail 发送密码重置邮件
func SendPasswordResetEmail(to, username, token string) error {
	config := hotdataserve.GetMailSettingsConfigCache()
	fromName, fromEmail := normalizeSender(config)
	slog.Debug("准备发送密码重置邮件", "to", to, "username", username, "enableMail", config.EnableMail, "smtpHost", config.SmtpHost, "smtpPort", config.SmtpPort, "fromName", fromName, "fromEmail", fromEmail)
	if !config.EnableMail {
		return errors.New("mail settings config is disabled")
	}
	message := mail.NewMsg()
	if err := message.To(to); err != nil {
		return fmt.Errorf("failed to set To address: %s", err)
	}
	message.Subject("密码重置请求")
	if err := setMessageFrom(message, config); err != nil {
		return err
	}
	body, err := generatePasswordResetEmailBody(username, token)
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
		slog.Debug("密码重置邮件 SMTP 发送失败", "to", to, "username", username, "err", err)
		return fmt.Errorf("failed to send mail: %s", err)
	}
	slog.Debug("密码重置邮件 SMTP 发送成功", "to", to, "username", username)
	return nil
}

// SendTestEmailWithConfig 使用指定配置发送测试邮件
func SendTestEmailWithConfig(config pageConfig.MailSettingsConfig, testEmail string) error {
	// 使用 go-mail 库直接发送测试邮件
	fromName, fromEmail := normalizeSender(config)
	slog.Debug("准备发送测试邮件", "to", testEmail, "enableMail", config.EnableMail, "smtpHost", config.SmtpHost, "smtpPort", config.SmtpPort, "fromName", fromName, "fromEmail", fromEmail)
	message := mail.NewMsg()
	if err := setMessageFrom(message, config); err != nil {
		return err
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
		slog.Debug("测试邮件 SMTP 发送失败", "to", testEmail, "err", err)
		return fmt.Errorf("failed to send mail: %s", err)
	}
	slog.Debug("测试邮件 SMTP 发送成功", "to", testEmail)

	return nil
}
