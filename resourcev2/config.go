package resourcev2

import (
	"embed"
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/spf13/cast"
	"html/template"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

//go:embed  all:templates/**
var templates embed.FS

func GetTemplates() *template.Template {
	tmpl := template.New("resource_v2").
		Funcs(template.FuncMap{
			"ContainsInt": func(s []int, v any) bool {
				return slices.Contains(s, cast.ToInt(v))
			},
			"GetMetaList":       GetMetaList,
			"GetRealFilePath":   GetRealFilePath,
			"GetImportInfoPath": GetImportInfoPath,
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

// GetStaticFS 返回静态文件的文件系统
func GetStaticFS() (fs.FS, error) {
	if !setting.IsProduction() {
		return os.DirFS(filepath.Join("resourceV2", "static")), nil
	}
	static, err := fs.Sub(GetViewAssert(), "static")
	if err != nil {
		return nil, err
	}
	return static, nil
}

type MetaItem struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

func GetMetaList() []MetaItem {
	return jsonopt.Decode[[]MetaItem](preferences.Get("site.metaList", "[]"))
}

type ManifestItem struct {
	File    string   `json:"file"`
	Name    string   `json:"name"`
	Src     string   `json:"src"`
	IsEntry bool     `json:"isEntry"`
	Imports []string `json:"imports"`
	Css     []string `json:"css"`
}

var manifestMap = map[string]string{}
var manifestItemMap = map[string]ManifestItem{}

func init() {
	content, err := viewAssert.ReadFile(filepath.Join("static", "dist", ".vite", "manifest.json"))
	if err != nil {
		slog.Error("ManifestGetError")
		return
	}

	info := jsonopt.Decode[map[string]ManifestItem](content)
	manifestItemMap = info
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

var htmlHeaderCache sync.Map

func GetImportInfoPath(origin string) template.HTML {
	if item, ok := manifestItemMap[origin]; ok {
		if val, cached := htmlHeaderCache.Load(origin); cached {
			return val.(template.HTML)
		}

		sb := &strings.Builder{}

		// 根据文件扩展名决定加载方式
		fileExt := filepath.Ext(item.File)
		switch fileExt {
		case ".js", ".mjs", ".ts", ".jsx", ".tsx":
			sb.WriteString(fmt.Sprintf(`<script type="module" src="/%s"></script>`, item.File))
		case ".css":
			sb.WriteString(fmt.Sprintf(`<link rel="stylesheet" href="/%s">`, item.File))
		default:
			// 非标准文件类型使用原处理逻辑
			sb.WriteString(fmt.Sprintf(`<script type="module" src="/%s"></script>`, item.File))
		}
		sb.WriteByte('\n')

		// 单独处理关联的CSS文件
		for _, css := range item.Css {
			if css != item.File { // 避免重复添加
				sb.WriteString(fmt.Sprintf(`<link rel="stylesheet" href="/%s">`, css))
				sb.WriteByte('\n')
			}
		}

		res := template.HTML(sb.String())
		htmlHeaderCache.Store(origin, res)
		return res
	}
	return ""
}
