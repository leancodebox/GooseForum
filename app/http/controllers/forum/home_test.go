package forum

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
)

func TestHomePageRequestReturnsPayload(t *testing.T) {
	ensureForumTestAccessCategory(t, 994001)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/", Home)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Goose-Page", "true")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("Content-Type") == "" {
		t.Fatal("expected JSON content type")
	}
	if got := recorder.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("expected page payload response to avoid browser cache, got %q", got)
	}
}

func TestHomeHTMLReturnsNoJSContent(t *testing.T) {
	ensureForumTestAccessCategory(t, 994002)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/", Home)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `id="goose-app"`) {
		t.Fatalf("expected app mount point in HTML: %s", body)
	}
	if !strings.Contains(body, `id="goose-payload"`) {
		t.Fatalf("expected initial payload in HTML: %s", body)
	}
	if !strings.Contains(body, `<noscript>`) {
		t.Fatalf("expected noscript fallback in HTML: %s", body)
	}
	if strings.Contains(body, `goose-seo-content`) {
		t.Fatalf("expected no hidden SEO duplicate in HTML: %s", body)
	}
}

func TestHomeSidebarActiveKeyMatchesSelectedSort(t *testing.T) {
	for _, test := range []struct{ sort, key string }{
		{"latest", "topics"}, {"hot", "hot"}, {"popular", "popular"}, {"unknown", "topics"},
	} {
		if got := activeKeyForHome(test.sort); got != test.key {
			t.Errorf("sort %q: expected sidebar key %q, got %q", test.sort, test.key, got)
		}
	}
}

func TestReactResourceEntryIncludesRefreshPreamble(t *testing.T) {
	html := string(resourceEntry("site"))
	if setting.IsProduction() {
		if !strings.Contains(html, "/assets/react/") || strings.Contains(html, "@vite/client") {
			t.Fatalf("invalid React production entry: %s", html)
		}
		return
	}
	if !strings.Contains(html, "__vite_plugin_react_preamble_installed__") || !strings.Contains(html, "/src/site/main.tsx") {
		t.Fatalf("missing React development entry: %s", html)
	}
}

func TestReactProductionManifestEntry(t *testing.T) {
	entries := map[string]manifestItem{
		"index.html": {File: "assets/site.js", Imports: []string{"shared"}},
		"shared":     {Css: []string{"assets/app.css"}},
	}
	html := string(manifestEntry(entries, "index.html", "react/"))
	for _, path := range []string{"/assets/react/assets/site.js", "/assets/react/assets/app.css"} {
		if !strings.Contains(html, path) {
			t.Fatalf("missing production resource %s: %s", path, html)
		}
	}
	if strings.Contains(html, "@vite") {
		t.Fatal("production entry contains Vite development script")
	}
}

func TestResourceAssetUsesConfiguredCDN(t *testing.T) {
	old := preferences.GetString("app.cdn_url", "")
	preferences.Set("app.cdn_url", "https://cdn.example.com/forum/")
	t.Cleanup(func() { preferences.Set("app.cdn_url", old) })

	if got := resourceAsset("react/assets/site.js"); got != "https://cdn.example.com/forum/assets/react/assets/site.js" {
		t.Fatalf("resource asset = %q", got)
	}
}

func TestAdminResourceEntryUsesIndependentReactEntry(t *testing.T) {
	html := string(resourceEntry("admin"))
	if setting.IsProduction() {
		item := reactManifest["admin/index.html"]
		if item.File == "" || !strings.Contains(html, "/assets/react/"+item.File) {
			t.Fatalf("missing React admin production entry: %s", html)
		}
	} else if !strings.Contains(html, "/src/admin/main.tsx") || !strings.Contains(html, "__vite_plugin_react_preamble_installed__") {
		t.Fatalf("missing React admin development entry: %s", html)
	}
	if strings.Contains(html, "/src/site/main.tsx") {
		t.Fatalf("admin must not load site entry: %s", html)
	}
}
