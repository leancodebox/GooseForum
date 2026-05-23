package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/resource"
)

func TestAssetsAreGzipped(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type manifestItem struct {
		File string   `json:"file"`
		Css  []string `json:"css"`
	}

	content, err := resource.GetTemplateFS().Open("static/dist/.vite/manifest.json")
	if err != nil {
		t.Fatalf("open manifest: %v", err)
	}
	defer content.Close()

	var manifest map[string]manifestItem
	if err := json.NewDecoder(content).Decode(&manifest); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}

	entry := manifest["src/site/main.ts"]
	targets := []string{entry.File}
	targets = append(targets, entry.Css...)
	for _, target := range targets {
		if target == "" {
			t.Fatal("manifest contains empty asset path")
		}
		assertGzipAsset(t, "/assets/"+strings.TrimPrefix(target, "/"))
	}
}

func assertGzipAsset(t *testing.T, path string) {
	t.Helper()

	router := gin.New()
	assertRouter(router)

	for i := 0; i < 2; i++ {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.Header.Set("Accept-Encoding", "gzip")
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("%s request %d status = %d, want 200", path, i+1, response.Code)
		}
		if got := response.Header().Get("Content-Encoding"); got != "gzip" {
			t.Fatalf("%s request %d Content-Encoding = %q, want gzip", path, i+1, got)
		}
		if response.Body.Len() == 0 {
			t.Fatalf("%s request %d body is empty", path, i+1)
		}
	}
}
