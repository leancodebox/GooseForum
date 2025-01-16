package views

import (
	"embed"
	"github.com/leancodebox/GooseForum/app/bundles/goose/preferences"
	"html/template"
	"io/fs"
)

//go:embed *.html
var templateFiles embed.FS

// 添加模板函数映射
var templateFuncs = template.FuncMap{
	"getFooterLink": GetFooterLink,
}

// getFooterLink 获取页脚链接
func GetFooterLink() map[string]string {
	return map[string]string{
		"url":  preferences.Get("footer.url", "/"),
		"text": preferences.Get("footer.text", "Goos"),
	}
}

// GetTemplates 获取所有模板
func GetTemplates() (*template.Template, error) {
	// 创建带有函数的模板
	tmpl := template.New("").Funcs(template.FuncMap{
		"getFooterLink": GetFooterLink,
	})

	err := fs.WalkDir(templateFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			content, err := templateFiles.ReadFile(path)
			if err != nil {
				return err
			}

			_, err = tmpl.New(path).Parse(string(content))
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return tmpl, nil
}
