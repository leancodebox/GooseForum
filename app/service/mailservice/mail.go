package mailservice

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"html/template"
	"mime"
	"net/smtp"
	"time"
)

type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	FromName string
}

func getEmailConfig() EmailConfig {
	return EmailConfig{
		Host:     preferences.GetString("mail.host"),
		Port:     preferences.GetInt("mail.port"),
		Username: preferences.GetString("mail.username"),
		Password: preferences.GetString("mail.password"),
		FromName: preferences.GetString("mail.from_name"),
	}
}

// 添加一个用于发送邮件的客户端结构体
type emailClient struct {
	config EmailConfig
	client *smtp.Client
}

// 创建新的邮件客户端
func newEmailClient(config EmailConfig) (*emailClient, error) {
	// 创建TLS配置
	tlsConfig := &tls.Config{
		ServerName:         config.Host,
		InsecureSkipVerify: false, // 在生产环境中应该设置为 false
	}

	// 建立TLS连接
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port), tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("连接到SMTP服务器失败: %v", err)
	}

	// 创建SMTP客户端
	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		return nil, fmt.Errorf("创建SMTP客户端失败: %v", err)
	}

	// 进行身份验证
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
	if err = client.Auth(auth); err != nil {
		client.Close()
		return nil, fmt.Errorf("SMTP认证失败: %v", err)
	}

	return &emailClient{
		config: config,
		client: client,
	}, nil
}

// 关闭客户端连接
func (c *emailClient) close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// 发送邮件的方法
func (c *emailClient) send(to string, subject string, body string) error {
	// 设置发件人
	if err := c.client.Mail(c.config.Username); err != nil {
		return fmt.Errorf("设置发件人失败: %v", err)
	}

	// 设置收件人
	if err := c.client.Rcpt(to); err != nil {
		return fmt.Errorf("设置收件人失败: %v", err)
	}

	// 构建邮件头
	from := mime.QEncoding.Encode("utf-8", c.config.FromName) + " <" + c.config.Username + ">"
	date := time.Now().Format(time.RFC1123Z)

	header := make(map[string]string)
	header["From"] = from
	header["To"] = to
	header["Subject"] = mime.QEncoding.Encode("utf-8", subject)
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=UTF-8"
	header["Content-Transfer-Encoding"] = "base64"
	header["Date"] = date

	// 构建完整的邮件内容
	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	// 发送邮件内容
	w, err := c.client.Data()
	if err != nil {
		return fmt.Errorf("创建邮件内容写入器失败: %v", err)
	}
	defer w.Close()

	_, err = w.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("写入邮件内容失败: %v", err)
	}

	return nil
}

func SendActivationEmail(to, username, token string) error {
	config := getEmailConfig()

	// 创建邮件客户端
	client, err := newEmailClient(config)
	if err != nil {
		return fmt.Errorf("创建邮件客户端失败: %v", err)
	}
	defer client.close()

	// 生成邮件内容
	body, err := generateActivationEmailBody(username, token)
	if err != nil {
		return fmt.Errorf("生成邮件内容失败: %v", err)
	}

	// 发送邮件
	return client.send(to, "账号激活", body)
}

func generateActivationEmailBody(username, token string) (string, error) {
	// 邮件模板
	const emailTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>账号激活</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .container {
            background-color: #f9f9f9;
            border-radius: 5px;
            padding: 20px;
            margin-top: 20px;
        }
        .button {
            display: inline-block;
            padding: 10px 20px;
            background-color: #4CAF50;
            color: white;
            text-decoration: none;
            border-radius: 3px;
            margin: 20px 0;
        }
        .note {
            font-size: 0.9em;
            color: #666;
            margin-top: 20px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h2>欢迎加入 GooseForum</h2>
        <p>亲爱的 {{.Username}}:</p>
        <p>请点击以下按钮激活您的账号:</p>
        <a href="{{.ActivationLink}}" class="button">激活账号</a>
        <p class="note">如果按钮无法点击，请复制以下链接到浏览器打开：</p>
        <p>{{.ActivationLink}}</p>
        <p class="note">此链接24小时内有效</p>
    </div>
</body>
</html>
`

	// 解析并执行模板
	tmpl, err := template.New("activation").Parse(emailTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]any{
		"Username": username,
		"ActivationLink": fmt.Sprintf("%s/api/activate?token=%s",
			preferences.GetString("server.url"), token),
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
