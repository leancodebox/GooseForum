package resource

import (
	"embed"
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
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
	return !setting.IsProduction()
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

type MetaItem struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

func GetMetaList() []MetaItem {
	return jsonopt.Decode[[]MetaItem](preferences.Get("site.metaList", "[]"))
}

func Iterate(count int) []int {
	var items []int
	for i := 0; i < count; i++ {
		items = append(items, i)
	}
	return items
}

// GetTemplates 返回所有模板
func GetTemplates() *template.Template {
	tmpl := template.New("root").Funcs(template.FuncMap{
		"Iterate":       Iterate,
		"getFooterLink": GetFooterLink,
		"metaList":      GetMetaList,
	})

	if isDevelopment() {
		fmt.Println("开发模式")
		// 开发模式下直接从目录读取模板
		return template.Must(template.Must(tmpl.ParseGlob(filepath.Join("resource", "templates", "*.gohtml"))).
			ParseGlob(filepath.Join("resource", "templates", "*", "*.gohtml")))
	}

	// 生产模式下从embed读取模板
	return template.Must(tmpl.ParseFS(templatesFS,
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
