package routes

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/resource"
)

func TestPublicSiteStatusAndLegacyUnreadAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	apiRoute(router)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest("GET", "/api/forum/status", nil))
	if recorder.Code != 200 {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatal("status must not be cached")
	}
	var body struct {
		Result map[string]any `json:"result"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Result["deploymentId"] != resource.DeploymentID() {
		t.Fatalf("deployment: %+v", body)
	}
	if _, ok := body.Result["unread"]; ok {
		t.Fatal("guest must not receive personal status")
	}
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest("GET", "/api/forum/unread-status", nil))
	if recorder.Code != 401 {
		t.Fatalf("legacy unread auth=%d", recorder.Code)
	}
}
