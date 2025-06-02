package resourcev2

import (
	"embed"
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/spf13/cast"
	"html/template"
	"path/filepath"
	"slices"
)

//go:embed  all:templates/**
var templates embed.FS

func GetTemplates() *template.Template {
	tmpl := template.New("resource_v2").
		Funcs(template.FuncMap{
			"containsInt": func(s []int, v any) bool {
				return slices.Contains(s, cast.ToInt(v))
			},
		})
	if !setting.IsProduction() {
		fmt.Println("开发模式")
		// 开发模式下直接从目录读取模板
		return template.Must(template.Must(
			tmpl.ParseGlob(filepath.Join("resourcev2", "templates", "*.gohtml"))).
			ParseGlob(filepath.Join("resourcev2", "templates", "*.gohtml")))
	}
	return template.Must(tmpl.ParseFS(templates,
		"templates/*.gohtml",
	))
}

//go:embed all:static/**
var viewAssert embed.FS

func GetViewAssert() *embed.FS {
	return &viewAssert
}
