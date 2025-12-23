package resource

import (
	"embed"
	"io/fs"
	"os"
	"path"

	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
)

// ThemeConfig represents the theme configuration
type ThemeConfig struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	Version          string `json:"version"`
	ViteDevServerURL string `json:"ViteDevServerURL"`
	Manifest         string `json:"manifest"`
}

//go:embed  all:templates/**
var templates embed.FS

//go:embed all:static/**
var viewAssert embed.FS

//go:embed templates/goose.theme.json
var themeConfig []byte

// GetTemplateFS returns the filesystem for templates
func GetTemplateFS() fs.FS {
	if !setting.IsProduction() {
		// In development mode, use the local file system
		return os.DirFS("resource")
	}
	return templates
}

func GetViewAssert() *embed.FS {
	return &viewAssert
}

// GetAssetsFS returns the filesystem for assets
func GetAssetsFS() (fs.FS, error) {
	if !setting.IsProduction() {
		return os.DirFS(path.Join("resource", "static", "dist", "assets")), nil
	}
	static, err := fs.Sub(GetViewAssert(), path.Join("static", "dist", "assets"))
	if err != nil {
		return nil, err
	}
	return static, nil
}

// GetStaticFS returns the filesystem for static files
func GetStaticFS() (fs.FS, error) {
	if !setting.IsProduction() {
		return os.DirFS(path.Join("resource", "static")), nil
	}
	static, err := fs.Sub(GetViewAssert(), "static")
	if err != nil {
		return nil, err
	}
	return static, nil
}

// GetThemeConfig returns the parsed theme configuration
func GetThemeConfig() (*ThemeConfig, error) {
	if len(themeConfig) == 0 {
		return nil, nil
	}
	return jsonopt.Decode[*ThemeConfig](themeConfig), nil
}
