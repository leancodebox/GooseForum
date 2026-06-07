package routes

import (
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/resource"
)

func TestAssetsGzipSwitch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type manifestItem struct {
		File string   `json:"file"`
		Css  []string `json:"css"`
	}

	content, err := resource.GetTemplateFS().Open("static/dist/.vite/manifest.json")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			t.Skip("vite manifest is missing; run pnpm -C resource build to enable asset gzip checks")
		}
		t.Fatalf("open manifest: %v", err)
	}
	defer func() { _ = content.Close() }()

	var manifest map[string]manifestItem
	if err := json.NewDecoder(content).Decode(&manifest); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}

	entry := manifest["src/site/main.ts"]
	targets := append([]string{entry.File}, entry.Css...)
	for _, target := range targets {
		if target == "" {
			t.Fatal("manifest contains empty asset path")
		}
		assertAssetGzip(t, "/assets/"+strings.TrimPrefix(target, "/"), true)
		assertAssetGzip(t, "/assets/"+strings.TrimPrefix(target, "/"), false)
	}
}

func assertAssetGzip(t *testing.T, path string, enabled bool) {
	t.Helper()

	previous := gzipEnabled()
	preferences.Set("server.gzip", enabled)
	t.Cleanup(func() {
		preferences.Set("server.gzip", previous)
	})

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
		got := response.Header().Get("Content-Encoding")
		if enabled && got != "gzip" {
			t.Fatalf("%s request %d Content-Encoding = %q, want gzip", path, i+1, got)
		}
		if !enabled && got == "gzip" {
			t.Fatalf("%s request %d Content-Encoding = gzip, want empty", path, i+1)
		}
		if response.Body.Len() == 0 {
			t.Fatalf("%s request %d body is empty", path, i+1)
		}
	}
}
