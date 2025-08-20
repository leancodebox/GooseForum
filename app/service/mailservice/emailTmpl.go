package mailservice

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"

	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
)

//go:embed activation-email.gohtml
var emailTemplate string
var emailTmpl *template.Template

//go:embed password-reset-email.gohtml
var passwordResetTemplate string
var passwordResetTmpl *template.Template

func generateActivationEmailBody(username, token string) (string, error) {
	if emailTmpl == nil {
		// 解析并执行模板
		tmpl, err := template.New("activation").Parse(emailTemplate)
		if err != nil {
			return "", err
		}
		emailTmpl = tmpl
	}
	var buf bytes.Buffer
	err := emailTmpl.Execute(&buf, map[string]any{
		"SiteName": hotdataserve.GetSiteSettingsConfigCache().SiteName,
		"Username": username,
		"ActivationLink": fmt.Sprintf("%s/api/activate?token=%s",
			preferences.GetString("server.url"), token),
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func generatePasswordResetEmailBody(username, token string) (string, error) {
	if passwordResetTmpl == nil {
		// 解析并执行模板
		tmpl, err := template.New("passwordReset").Parse(passwordResetTemplate)
		if err != nil {
			return "", err
		}
		passwordResetTmpl = tmpl
	}
	var buf bytes.Buffer
	err := passwordResetTmpl.Execute(&buf, map[string]any{
		"Username": username,
		"ResetLink": fmt.Sprintf("%s/reset-password?token=%s",
			preferences.GetString("server.url"), token),
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
