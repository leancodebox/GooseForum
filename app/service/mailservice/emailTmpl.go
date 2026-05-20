package mailservice

import (
	"bytes"
	_ "embed"
	"html/template"
	"net/url"
	"strings"

	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"
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
		"SiteName":       siteConfig.SiteName,
		"Username":       username,
		"ActivationLink": buildEmailActionURL(emailSiteBaseURL(siteConfig.SiteUrl), urlconfig.Activate(), token),
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
		"Username":  username,
		"ResetLink": buildEmailActionURL(emailSiteBaseURL(siteConfig.SiteUrl), urlconfig.ResetPassword(), token),
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func emailSiteBaseURL(siteURL string) string {
	baseURL := strings.TrimSpace(siteURL)
	if baseURL == "" {
		baseURL = strings.TrimSpace(preferences.GetString("server.url", ""))
	}
	return strings.TrimRight(baseURL, "/")
}

func buildEmailActionURL(baseURL, actionPath, token string) string {
	cleanBaseURL := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	cleanPath := "/" + strings.TrimLeft(actionPath, "/")
	actionURL := cleanBaseURL + cleanPath
	query := url.Values{}
	query.Set("token", token)
	return actionURL + "?" + query.Encode()
}
