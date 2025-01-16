package component

import (
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"html/template"
	"sync"

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
			templates, err = template.ParseGlob("app/views/*.html")
		}
		if err != nil {
			panic(err)
		}
	})
	return templates
}
