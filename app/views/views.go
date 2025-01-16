package views

import (
	"embed"
	"html/template"
	"io/fs"
)

//go:embed *.html
var templateFiles embed.FS

// GetTemplates 获取所有模板
func GetTemplates() (*template.Template, error) {
	// 读取所有模板文件
	tmpl := template.New("")

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
