package mailservice

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/i18n"
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
		return fmt.Errorf("failed to set From address: %w", err)
	}
	return nil
}

// CheckConfigured validates local mail settings without connecting to SMTP.
func CheckConfigured() error {
	return validateConfiguration(hotdataserve.GetMailSettingsConfigCache())
}

func validateConfiguration(config pageConfig.MailSettingsConfig) error {
	if !config.EnableMail {
		return errors.New("mail settings config is disabled")
	}
	if strings.TrimSpace(config.SmtpHost) == "" || config.SmtpPort < 1 || config.SmtpPort > 65535 {
		return errors.New("SMTP host or port is invalid")
	}
	if err := setMessageFrom(mail.NewMsg(), config); err != nil {
		return err
	}
	_, err := buildClientByConfig(config)
	return err
}

func SendActivationEmail(to, username, token string, locale ...string) error {
	body, err := generateActivationEmailBody(username, token, locale...)
	if err != nil {
		return fmt.Errorf("生成邮件内容失败: %w", err)
	}
	return sendTemplatedEmail("激活", to, username, i18n.T(emailBodyLang(locale...), "email.activation.subject"), body)
}

// SendPasswordResetEmail 发送密码重置邮件
func SendPasswordResetEmail(to, username, token string, locale ...string) error {
	body, err := generatePasswordResetEmailBody(username, token, locale...)
	if err != nil {
		return fmt.Errorf("生成邮件内容失败: %w", err)
	}
	return sendTemplatedEmail("密码重置", to, username, i18n.T(emailBodyLang(locale...), "email.passwordReset.subject"), body)
}

// SendRegistrationInvitationEmail sends an invitation link for the public registration form.
func SendRegistrationInvitationEmail(to, code string, locale ...string) error {
	body, err := generateRegistrationInvitationEmailBody(code, locale...)
	if err != nil {
		return fmt.Errorf("生成邮件内容失败: %w", err)
	}
	return sendTemplatedEmail("注册邀请", to, "", i18n.T(emailBodyLang(locale...), "email.registrationInvitation.subject"), body)
}

func sendTemplatedEmail(kind, to, username, subject, body string) error {
	config := hotdataserve.GetMailSettingsConfigCache()
	fromName, fromEmail := normalizeSender(config)
	slog.Debug("准备发送邮件", "kind", kind, "to", to, "username", username, "enableMail", config.EnableMail, "smtpHost", config.SmtpHost, "smtpPort", config.SmtpPort, "fromName", fromName, "fromEmail", fromEmail)
	if !config.EnableMail {
		return errors.New("mail settings config is disabled")
	}
	message := mail.NewMsg()
	if err := message.To(to); err != nil {
		return fmt.Errorf("failed to set To address: %w", err)
	}
	message.Subject(subject)
	if err := setMessageFrom(message, config); err != nil {
		return err
	}
	message.SetBodyString(mail.TypeTextHTML, body)

	client, err := buildClientByConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create mail client: %w", err)
	}
	defer func() { _ = client.Close() }()
	if err = client.DialAndSend(message); err != nil {
		slog.Debug("邮件 SMTP 发送失败", "kind", kind, "to", to, "username", username, "err", err)
		return fmt.Errorf("failed to send mail: %w", err)
	}
	slog.Debug("邮件 SMTP 发送成功", "kind", kind, "to", to, "username", username)
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
		return fmt.Errorf("failed to set To address: %w", err)
	}
	message.Subject("邮件配置测试")
	message.SetBodyString(mail.TypeTextPlain,
		fmt.Sprintf("[%v]这是一封测试邮件,用于验证邮件配置是否正确。", time.Now().Format(time.DateTime)))

	client, err := buildClientByConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create mail client: %w", err)
	}
	defer func() { _ = client.Close() }()

	if err = client.DialAndSend(message); err != nil {
		slog.Debug("测试邮件 SMTP 发送失败", "to", testEmail, "err", err)
		return fmt.Errorf("failed to send mail: %w", err)
	}
	slog.Debug("测试邮件 SMTP 发送成功", "to", testEmail)

	return nil
}
