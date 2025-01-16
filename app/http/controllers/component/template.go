package component

import (
	"html/template"
	"sync"

	"github.com/leancodebox/GooseForum/app/bundles/setting"

	"github.com/leancodebox/GooseForum/app/views"
)

var (
	templates     *template.Template
	templatesOnce sync.Once
)

// GetTemplates 获取模板单例
func GetTemplates() *template.Template {
	templatesOnce.Do(func() {
		var err error
		if setting.IsProduction() {
			templates, err = views.GetTemplates()
		} else {
			// 创建带有函数的模板
			templates = template.New("").Funcs(template.FuncMap{
				"getFooterLink": views.GetFooterLink,
			})
			templates, err = templates.ParseGlob("app/views/*.html")
		}
		if err != nil {
			panic(err)
		}
	})
	return templates
}
