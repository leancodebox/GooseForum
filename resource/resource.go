package resource

import (
	"embed"
	"github.com/leancodebox/GooseForum/app/bundles/goose/preferences"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed templates/*
var templatesFS embed.FS

func GetTemplatesFS() embed.FS {
	return templatesFS
}

// isDevelopment 判断是否为开发模式
func isDevelopment() bool {
	return preferences.Get("app.env", "production") == "local"
}

//go:embed static/*
var staticFS embed.FS

// GetFooterLink 获取页脚链接
func GetFooterLink() map[string]string {
	return map[string]string{
		"url":  preferences.Get("footer.url", "/"),
		"text": preferences.Get("footer.text", "Goos"),
	}
}

// GetTemplates 返回所有模板
func GetTemplates() *template.Template {
	return template.Must(template.New("root").Funcs(template.FuncMap{
		"getFooterLink": GetFooterLink,
	}).ParseFS(templatesFS,
		"templates/*.gohtml",
		"templates/*/**.gohtml",
	))
}

// GetStaticFS 返回静态文件的文件系统
func GetStaticFS() (fs.FS, error) {
	if isDevelopment() {
		return os.DirFS(filepath.Join("resource", "static")), nil
	}
	static, err := fs.Sub(staticFS, "static")
	if err != nil {
		return nil, err
	}
	return static, nil
}
