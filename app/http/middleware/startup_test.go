package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestStartupGateBlocksUntilComplete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gate := NewStartupGate()
	router := gin.New()
	router.Use(gate.Handler)
	router.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "ready") })

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable || response.Header().Get("Cache-Control") != "no-store, no-cache, must-revalidate" || !strings.Contains(response.Body.String(), "class=\"signal\"") {
		t.Fatalf("pending response = %d %q", response.Code, response.Body.String())
	}

	gate.Complete()
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Body.String() != "ready" {
		t.Fatalf("ready response = %d %q", response.Code, response.Body.String())
	}
}

func TestStartupGateRemainsPendingUntilComplete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gate := NewStartupGate()
	router := gin.New()
	router.Use(gate.Handler)
	router.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "ready") })
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusServiceUnavailable || !strings.Contains(response.Body.String(), "class=\"signal\"") {
		t.Fatalf("pending response = %d %q", response.Code, response.Body.String())
	}
}
