package viewrender

import (
	"strings"
	"testing"
)

func TestViteHandler_DevMode(t *testing.T) {
	devServerURL := "http://localhost:3009"
	handler := NewViteHandler(devServerURL, false, nil)

	entry := "src/main.ts"
	html := handler.ViteEntry(entry)
	expected := `<script type="module" src="http://localhost:3009/src/main.ts"></script>`
	if !strings.Contains(string(html), expected) {
		t.Errorf("Expected HTML to contain %s, got %s", expected, html)
	}

	path := handler.VitePath(entry)
	expectedPath := "http://localhost:3009/src/main.ts"
	if path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, path)
	}
}

func TestViteHandler_ProdMode(t *testing.T) {
	manifest := map[string]ManifestItem{
		"src/main.ts": {
			File:    "assets/main.123456.js",
			IsEntry: true,
			Css:     []string{"assets/style.123456.css"},
		},
	}
	handler := NewViteHandler("", true, manifest)

	// Test VitePath
	path := handler.VitePath("src/main.ts")
	
	if !strings.Contains(path, "assets/main.123456.js") {
		t.Errorf("Expected path to contain assets/main.123456.js, got %s", path)
	}

	// Test ViteEntry
	html := handler.ViteEntry("src/main.ts")
	htmlStr := string(html)

	if !strings.Contains(htmlStr, `src="/assets/main.123456.js"`) && !strings.Contains(htmlStr, `src="assets/main.123456.js"`) {
		t.Errorf("Expected HTML to contain src script, got %s", htmlStr)
	}
	if !strings.Contains(htmlStr, `href="/assets/style.123456.css"`) && !strings.Contains(htmlStr, `href="assets/style.123456.css"`) {
		t.Errorf("Expected HTML to contain css link, got %s", htmlStr)
	}
}
