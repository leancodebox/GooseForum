package controllers

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
	"unicode"

	"github.com/leancodebox/GooseForum/app/bundles/i18n"
	"github.com/leancodebox/GooseForum/resource"
)

// TestActivationPageLocalized renders the activation view with the English
// translator and confirms the labels localize with no residual Chinese.
func TestActivationPageLocalized(t *testing.T) {
	tmpl, err := template.ParseFS(resource.GetTemplateFS(), "templates/view/activate.gohtml")
	if err != nil {
		t.Fatalf("parse activation template: %v", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct {
		Data ActivateAccountData
		T    func(string, ...any) string
		Lang string
	}{
		Data: ActivateAccountData{
			Title:       i18n.T("en", "activationTitleSuccess"),
			Message:     i18n.T("en", "activationSuccess"),
			Success:     true,
			Description: i18n.T("en", "activationDescSuccess"),
		},
		T:    i18n.Func("en"),
		Lang: "en",
	})
	if err != nil {
		t.Fatalf("render activation template: %v", err)
	}

	out := buf.String()
	for _, want := range []string{`lang="en"`, "Log in now", "Back to home", "Completed", "Account activated"} {
		if !strings.Contains(out, want) {
			t.Errorf("activation page missing %q", want)
		}
	}
	for _, r := range out {
		if unicode.Is(unicode.Han, r) {
			t.Errorf("activation page still contains Chinese: %q", string(r))
			break
		}
	}
}
