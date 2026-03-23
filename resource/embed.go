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

//go:embed all:templates all:static all:locales
var resources embed.FS

// GetTemplateFS returns the filesystem for templates
func GetTemplateFS() fs.FS {
	if !setting.IsProduction() {
		// In development mode, use the local file system
		return os.DirFS("resource")
	}
	return resources
}

// GetAssetsFS returns the filesystem for assets
func GetAssetsFS() (fs.FS, error) {
	if !setting.IsProduction() {
		return os.DirFS(path.Join("resource", "static", "dist", "assets")), nil
	}
	static, err := fs.Sub(resources, path.Join("static", "dist", "assets"))
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
	static, err := fs.Sub(resources, "static")
	if err != nil {
		return nil, err
	}
	return static, nil
}

// GetAdminFS  返回静态文件的文件系统
func GetAdminFS() (fs.FS, error) {
	if !setting.IsProduction() {
		return os.DirFS(path.Join("resource", "static", "admin", "dict")), nil
	}
	return fs.Sub(resources, path.Join("static", "admin", "dict"))
}

// GetThemeConfig returns the parsed theme configuration
func GetThemeConfig() (*ThemeConfig, error) {
	data, err := fs.ReadFile(GetTemplateFS(), "templates/goose.theme.json")
	if err != nil {
		return nil, err
	}
	return jsonopt.Decode[*ThemeConfig](data), nil
}
