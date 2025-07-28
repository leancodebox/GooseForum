package mailservice

import "github.com/leancodebox/GooseForum/app/bundles/preferences"

type EmailConfig struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	FromName  string `json:"fromName"`
	FromEmail string `json:"fromEmail"`
}

func getEmailConfig() EmailConfig {
	return EmailConfig{
		Host:      preferences.GetString("mail.host"),
		Port:      preferences.GetInt("mail.port"),
		Username:  preferences.GetString("mail.username"),
		Password:  preferences.GetString("mail.password"),
		FromName:  preferences.GetString("mail.from_name"),
		FromEmail: preferences.GetString("mail.from_email"),
	}
}
