package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
)

func TestBrowserCacheProduction(t *testing.T) {
	oldEnv := preferences.GetString("app.env", "production")
	t.Cleanup(func() {
		preferences.Set("app.env", oldEnv)
	})

	preferences.Set("app.env", "production")
	recorder := requestWithMiddleware(BrowserCache, http.MethodGet)

	if got := recorder.Header().Get("Cache-Control"); got != "public, max-age=18144000" {
		t.Fatalf("Cache-Control = %q, want long public cache", got)
	}
}

func TestBrowserCacheLocal(t *testing.T) {
	oldEnv := preferences.GetString("app.env", "production")
	t.Cleanup(func() {
		preferences.Set("app.env", oldEnv)
	})

	preferences.Set("app.env", "local")
	recorder := requestWithMiddleware(BrowserCache, http.MethodGet)

	if got := recorder.Header().Get("Cache-Control"); got != "" {
		t.Fatalf("Cache-Control = %q, want empty local cache header", got)
	}
}

func TestGinCors(t *testing.T) {
	recorder := requestWithMiddleware(GinCors, http.MethodGet)

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want *", got)
	}
	if got := recorder.Header().Get("Access-Control-Expose-Headers"); got == "" {
		t.Fatalf("expected exposed CORS headers")
	}
}

func TestGinCorsOptions(t *testing.T) {
	recorder := requestWithMiddleware(GinCors, http.MethodOptions)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("OPTIONS status = %d, want 204", recorder.Code)
	}
}

func TestSiteInfo(t *testing.T) {
	recorder := requestWithMiddleware(SiteInfo, http.MethodGet)

	if got := recorder.Header().Get("X-Powered-By"); got != "GooseForum/0.0.1" {
		t.Fatalf("X-Powered-By = %q, want GooseForum/0.0.1", got)
	}
}

func requestWithMiddleware(middleware gin.HandlerFunc, method string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware)
	router.Handle(method, "/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, "/", nil)
	router.ServeHTTP(recorder, request)
	return recorder
}
