package resourcev2

import (
	"embed"
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/spf13/cast"
	"html/template"
	"log/slog"
	"path/filepath"
	"slices"
)

//go:embed  all:templates/**
var templates embed.FS

func GetTemplates() *template.Template {
	tmpl := template.New("resource_v2").
		Funcs(template.FuncMap{
			"ContainsInt": func(s []int, v any) bool {
				return slices.Contains(s, cast.ToInt(v))
			},
			"GetMetaList":     GetMetaList,
			"GetRealFilePath": GetRealFilePath,
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

type MetaItem struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

func GetMetaList() []MetaItem {
	return jsonopt.Decode[[]MetaItem](preferences.Get("site.metaList", "[]"))
}

var manifestMap = map[string]string{}

func init() {
	content, err := viewAssert.ReadFile(filepath.Join("static", "dist", ".vite", "manifest.json"))
	if err != nil {
		slog.Error("ManifestGetError")
		return
	}
	type ManifestItem struct {
		File    string   `json:"file"`
		Name    string   `json:"name"`
		Src     string   `json:"src"`
		IsEntry bool     `json:"isEntry"`
		Imports []string `json:"imports"`
		Css     []string `json:"css"`
	}
	info := jsonopt.Decode[map[string]ManifestItem](content)
	newManifestMap := map[string]string{}
	for s, item := range info {
		if item.File != "" {
			newManifestMap[s] = item.File
		}
	}
	manifestMap = newManifestMap
}

func GetRealFilePath(origin string) string {
	return manifestMap[origin]
}
