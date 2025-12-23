package mailservice

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"

	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
)

//go:embed activation-email.gohtml
var emailTemplate string
var emailTmpl *template.Template

//go:embed password-reset-email.gohtml
var passwordResetTemplate string
var passwordResetTmpl *template.Template

func init() {
	emailTmpl = template.Must(template.New("activation").Parse(emailTemplate))
	passwordResetTmpl = template.Must(template.New("passwordReset").Parse(passwordResetTemplate))
}

func generateActivationEmailBody(username, token string) (string, error) {
	siteConfig := hotdataserve.GetSiteSettingsConfigCache()
	var buf bytes.Buffer
	err := emailTmpl.Execute(&buf, map[string]any{
		"SiteName": hotdataserve.GetSiteSettingsConfigCache().SiteName,
		"Username": username,
		"ActivationLink": fmt.Sprintf("%s/activate?token=%s",
			siteConfig.SiteUrl, token),
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func generatePasswordResetEmailBody(username, token string) (string, error) {
	siteConfig := hotdataserve.GetSiteSettingsConfigCache()
	var buf bytes.Buffer
	err := passwordResetTmpl.Execute(&buf, map[string]any{
		"Username": username,
		"ResetLink": fmt.Sprintf("%s/reset-password?token=%s",
			siteConfig.SiteUrl, token),
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
