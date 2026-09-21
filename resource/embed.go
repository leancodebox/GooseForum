// Package resource exposes embedded templates and static assets.
package resource

import (
	"embed"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/leancodebox/GooseForum/app/bundles/setting"
)

// ThemeConfig describes the front-end resource bundle metadata.
type ThemeConfig struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	Version          string `json:"version"`
	ViteDevServerURL string `json:"ViteDevServerURL"`
	Manifest         string `json:"manifest"`
}

//go:embed all:templates all:static
var resources embed.FS

// GetTemplateFS returns the template filesystem for the current environment.
func GetTemplateFS() fs.FS {
	if !setting.IsProduction() {
		return os.DirFS(resourceDir())
	}
	return resources
}

// GetAssetsFS returns the built asset filesystem for the current environment.
func GetAssetsFS() (fs.FS, error) {
	if !setting.IsProduction() {
		return os.DirFS(filepath.Join(resourceDir(), "static", "dist")), nil
	}
	return fs.Sub(resources, path.Join("static", "dist"))
}

// GetStaticFS returns the static asset filesystem for the current environment.
func GetStaticFS() (fs.FS, error) {
	if !setting.IsProduction() {
		return os.DirFS(filepath.Join(resourceDir(), "static")), nil
	}
	return fs.Sub(resources, "static")
}

func resourceDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "resource"
	}
	for {
		candidate := filepath.Join(wd, "resource")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return "resource"
		}
		wd = parent
	}
}
