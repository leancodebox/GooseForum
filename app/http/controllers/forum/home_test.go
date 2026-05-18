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
