package mailservice

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"html/template"
)

//go:embed activation-email.gohtml
var emailTemplate string
var emailTmpl *template.Template

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
		"Username": username,
		"ActivationLink": fmt.Sprintf("%s/api/activate?token=%s",
			preferences.GetString("server.url"), token),
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
