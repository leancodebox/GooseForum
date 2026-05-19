package forum

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHomePageRequestReturnsPayload(t *testing.T) {
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
}

func TestHomeHTMLReturnsNoJSContent(t *testing.T) {
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

func TestResourceEntryUsesViteInLocalMode(t *testing.T) {
	if got := viteDevServerFor("", false); got != "http://localhost:3010" {
		t.Fatalf("expected local mode to default to Vite dev server, got %q", got)
	}
	if got := viteDevServerFor("http://127.0.0.1:4173", true); got != "http://127.0.0.1:4173" {
		t.Fatalf("expected explicit dev server to win, got %q", got)
	}
	if got := viteDevServerFor("", true); got != "" {
		t.Fatalf("expected production mode without override to use manifest, got %q", got)
	}
}

func TestResourceEntryUsesConfiguredDevServer(t *testing.T) {
	t.Setenv("GOOSE_UNUSED", "keeps test isolated")
	html := string(resourceEntry("src/main.ts"))
	devServer := viteDevServer()
	if devServer == "" {
		t.Skip("current test config uses production manifest")
	}
	if !strings.Contains(html, strings.TrimRight(devServer, "/")+`/assets/@vite/client`) {
		t.Fatalf("expected resource entry to include Vite client: %s", html)
	}
	if !strings.Contains(html, strings.TrimRight(devServer, "/")+`/assets/src/main.ts`) {
		t.Fatalf("expected resource entry to include Vite entry: %s", html)
	}
	if strings.Contains(html, `/assets/assets/`) {
		t.Fatalf("expected dev resource entry not to use built manifest assets: %s", html)
	}
}
