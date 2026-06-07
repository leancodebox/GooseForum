package resource

import (
	"encoding/json"
	"errors"
	"io/fs"
	"testing"
)

func TestAssetsFSMatchesViteManifestPaths(t *testing.T) {
	type manifestItem struct {
		File string   `json:"file"`
		CSS  []string `json:"css"`
	}

	content, err := fs.ReadFile(GetTemplateFS(), "static/dist/.vite/manifest.json")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			t.Skip("vite manifest is missing; run pnpm -C resource build to enable asset path checks")
		}
		t.Fatalf("read manifest: %v", err)
	}

	var manifest map[string]manifestItem
	if err := json.Unmarshal(content, &manifest); err != nil {
		t.Fatalf("parse manifest: %v", err)
	}

	assetsFS, err := GetAssetsFS()
	if err != nil {
		t.Fatalf("get assets fs: %v", err)
	}

	for _, entryPath := range []string{"src/site/main.ts", "src/admin/main.ts"} {
		entry, ok := manifest[entryPath]
		if !ok {
			t.Fatalf("%s is missing from manifest", entryPath)
		}

		if _, err := fs.Stat(assetsFS, entry.File); err != nil {
			t.Fatalf("entry file %q should be readable from assets fs: %v", entry.File, err)
		}
		for _, css := range entry.CSS {
			if _, err := fs.Stat(assetsFS, css); err != nil {
				t.Fatalf("css file %q should be readable from assets fs: %v", css, err)
			}
		}
	}
}

func TestStaticFS(t *testing.T) {
	staticFS, err := GetStaticFS()
	if err != nil {
		t.Fatalf("GetStaticFS failed: %v", err)
	}
	if _, err := fs.Stat(staticFS, "pic/icon_300.webp"); err != nil {
		t.Fatalf("expected bundled icon to be readable: %v", err)
	}
}
