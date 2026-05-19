package resource

import (
	"embed"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
)

type ThemeConfig struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	Version          string `json:"version"`
	ViteDevServerURL string `json:"ViteDevServerURL"`
	Manifest         string `json:"manifest"`
}

//go:embed all:templates all:static
var resources embed.FS

func GetTemplateFS() fs.FS {
	if !setting.IsProduction() {
		return os.DirFS(resourceDir())
	}
	return resources
}

func GetAssetsFS() (fs.FS, error) {
	if !setting.IsProduction() {
		return os.DirFS(filepath.Join(resourceDir(), "static", "dist")), nil
	}
	return fs.Sub(resources, path.Join("static", "dist"))
}

func GetStaticFS() (fs.FS, error) {
	if !setting.IsProduction() {
		return os.DirFS(filepath.Join(resourceDir(), "static")), nil
	}
	return fs.Sub(resources, "static")
}

func GetAdminFS() (fs.FS, error) {
	if !setting.IsProduction() {
		return os.DirFS(filepath.Join(resourceDir(), "static", "admin", "dict")), nil
	}
	return fs.Sub(resources, path.Join("static", "admin", "dict"))
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

func GetThemeConfig() (*ThemeConfig, error) {
	data, err := fs.ReadFile(GetTemplateFS(), "templates/goose.theme.json")
	if err != nil {
		return nil, err
	}
	return jsonopt.Decode[*ThemeConfig](data), nil
}
