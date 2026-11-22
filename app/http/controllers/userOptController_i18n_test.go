package controllers

import (
	"bytes"
	"html"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/i18n"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
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
			SiteName:    "GooseForum",
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

func TestActivationPageUsesConfiguredSiteIdentity(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&pageConfig.Entity{}); err != nil {
		t.Fatal(err)
	}
	setConfig := func(key string, value any, clearCache func()) {
		t.Helper()
		previous := pageConfig.GetByPageType(key)
		t.Cleanup(func() {
			if previous.Id == 0 {
				if err := db.Where("page_type = ?", key).Delete(&pageConfig.Entity{}).Error; err != nil {
					t.Error(err)
				}
			} else if err := pageConfig.SaveConfig(key, previous.Config); err != nil {
				t.Error(err)
			}
			clearCache()
		})
		if err := pageConfig.SaveConfig(key, jsonopt.Encode(value)); err != nil {
			t.Fatal(err)
		}
		clearCache()
	}
	router := gin.New()
	router.GET("/activate", ActivateAccount)
	render := func() string {
		t.Helper()
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/activate?lang=en", nil))
		if response.Code != http.StatusOK {
			t.Fatalf("activation status = %d", response.Code)
		}
		if response.Header().Get("Referrer-Policy") != "no-referrer" {
			t.Error("activation page should omit referrers for its configured assets")
		}
		return html.UnescapeString(response.Body.String())
	}

	setConfig(pageConfig.SiteSettings, pageConfig.SiteSettingsConfig{SiteName: "Example Forum", SiteEmail: "support+forum@example.test"}, hotdataserve.ClearSiteSettingsConfigCache)
	setConfig(pageConfig.SiteChrome, pageConfig.SiteChromeConfig{BrandType: "image", BrandImage: "/file/logo.png"}, hotdataserve.ClearSiteChromeConfigCache)
	body := render()
	for _, want := range []string{` - Example Forum</title>`, `href="/favicon.ico"`, `src="/file/logo.png" alt="Example Forum"`, `href="mailto:support+forum@example.test"`} {
		if !strings.Contains(body, want) {
			t.Errorf("activation page missing %q", want)
		}
	}

	setConfig(pageConfig.SiteSettings, pageConfig.SiteSettingsConfig{SiteName: "Example Forum"}, hotdataserve.ClearSiteSettingsConfigCache)
	setConfig(pageConfig.SiteChrome, pageConfig.SiteChromeConfig{BrandType: "text", BrandText: "Community"}, hotdataserve.ClearSiteChromeConfigCache)
	body = render()
	if !strings.Contains(body, "Community") || strings.Contains(body, "mailto:") {
		t.Fatal("activation page should show the configured text brand and omit an empty contact address")
	}
}
