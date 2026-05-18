package resource

import (
	"encoding/json"
	"io/fs"
	"testing"
)

func TestAssetsFSMatchesViteManifestPaths(t *testing.T) {
	type manifestItem struct {
		File string   `json:"file"`
		Css  []string `json:"css"`
	}

	content, err := fs.ReadFile(GetTemplateFS(), "static/dist/.vite/manifest.json")
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}

	var manifest map[string]manifestItem
	if err := json.Unmarshal(content, &manifest); err != nil {
		t.Fatalf("parse manifest: %v", err)
	}

	entry, ok := manifest["src/main.ts"]
	if !ok {
		t.Fatal("src/main.ts is missing from manifest")
	}

	assetsFS, err := GetAssetsFS()
	if err != nil {
		t.Fatalf("get assets fs: %v", err)
	}

	if _, err := fs.Stat(assetsFS, entry.File); err != nil {
		t.Fatalf("entry file %q should be readable from assets fs: %v", entry.File, err)
	}
	for _, css := range entry.Css {
		if _, err := fs.Stat(assetsFS, css); err != nil {
			t.Fatalf("css file %q should be readable from assets fs: %v", css, err)
		}
	}
}
