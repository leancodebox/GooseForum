package resource

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
			tmpl.ParseGlob(filepath.Join("resource", "templates", "*.gohtml"))).
			ParseGlob(filepath.Join("resource", "templates", "*", "*.gohtml")))
	}
	return template.Must(tmpl.ParseFS(templates,
		"templates/*.gohtml",
		"templates/*/*.gohtml",
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
		return os.DirFS(filepath.Join("resource", "static")), nil
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
	File             string   `json:"file"`
	Name             string   `json:"name"`
	Src              string   `json:"src"`
	IsEntry          bool     `json:"isEntry"`
	IsDynamicEntry   bool     `json:"isDynamicEntry"`
	Imports          []string `json:"imports"`
	DynamicImports   []string `json:"dynamicImports"`
	Css              []string `json:"css"`
	Assets           []string `json:"assets"`
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

// collectAllCSSFiles 递归收集所有依赖的 CSS 文件
func collectAllCSSFiles(entryKey string, visited map[string]bool) []string {
	if visited[entryKey] {
		return nil
	}
	visited[entryKey] = true

	var cssFiles []string
	if item, ok := manifestItemMap[entryKey]; ok {
		// 添加当前 chunk 的 CSS 文件
		cssFiles = append(cssFiles, item.Css...)

		// 递归处理所有 imports
		for _, importKey := range item.Imports {
			importedCSS := collectAllCSSFiles(importKey, visited)
			cssFiles = append(cssFiles, importedCSS...)
		}

		// 递归处理所有 dynamicImports
		for _, dynamicImportKey := range item.DynamicImports {
			dynamicCSS := collectAllCSSFiles(dynamicImportKey, visited)
			cssFiles = append(cssFiles, dynamicCSS...)
		}
	}
	return cssFiles
}

// collectAllModulePreloads 递归收集所有需要预加载的模块
func collectAllModulePreloads(entryKey string, visited map[string]bool) []string {
	if visited[entryKey] {
		return nil
	}
	visited[entryKey] = true

	var moduleFiles []string
	if item, ok := manifestItemMap[entryKey]; ok {
		// 递归处理所有 imports
		for _, importKey := range item.Imports {
			if importedItem, exists := manifestItemMap[importKey]; exists {
				// 添加导入的模块文件
				moduleFiles = append(moduleFiles, importedItem.File)
				// 递归处理嵌套的 imports
				nestedModules := collectAllModulePreloads(importKey, visited)
				moduleFiles = append(moduleFiles, nestedModules...)
			}
		}
	}
	return moduleFiles
}

// collectAllAssets 递归收集所有依赖的静态资源（字体、图片等）
func collectAllAssets(entryKey string, visited map[string]bool) []string {
	if visited[entryKey] {
		return nil
	}
	visited[entryKey] = true

	var assetFiles []string
	if item, ok := manifestItemMap[entryKey]; ok {
		// 添加当前 chunk 的静态资源
		assetFiles = append(assetFiles, item.Assets...)

		// 递归处理所有 imports
		for _, importKey := range item.Imports {
			importedAssets := collectAllAssets(importKey, visited)
			assetFiles = append(assetFiles, importedAssets...)
		}

		// 递归处理所有 dynamicImports
		for _, dynamicImportKey := range item.DynamicImports {
			dynamicAssets := collectAllAssets(dynamicImportKey, visited)
			assetFiles = append(assetFiles, dynamicAssets...)
		}
	}
	return assetFiles
}

func GetImportInfoPath(origin string) template.HTML {
	// 检查缓存
	cacheKey := fmt.Sprintf("%s_%v", origin, setting.IsProduction())
	if val, cached := htmlHeaderCache.Load(cacheKey); cached {
		return val.(template.HTML)
	}

	sb := &strings.Builder{}

	// 开发环境处理
	if !setting.IsProduction() {
		// 开发环境直接使用 Vite 开发服务器
		fileExt := filepath.Ext(origin)
		switch fileExt {
		case ".js", ".mjs", ".ts", ".jsx", ".tsx":
			sb.WriteString(fmt.Sprintf(`<script type="module" src="http://localhost:3001/%s"></script>`, origin))
		case ".css":
			sb.WriteString(fmt.Sprintf(`<link rel="stylesheet" href="http://localhost:3001/%s">`, origin))
		default:
			// 默认作为 JS 模块处理
			sb.WriteString(fmt.Sprintf(`<script type="module" src="http://localhost:3001/%s"></script>`, origin))
		}
		sb.WriteByte('\n')
	} else {
		// 生产环境使用 manifest.json 递归处理依赖
		if item, ok := manifestItemMap[origin]; ok {
			// 1. 递归收集所有 CSS 依赖
			visitedCSS := make(map[string]bool)
			allCSSFiles := collectAllCSSFiles(origin, visitedCSS)

			// 去重并添加 CSS 链接
			cssSet := make(map[string]bool)
			for _, css := range allCSSFiles {
				if !cssSet[css] {
					cssSet[css] = true
					sb.WriteString(fmt.Sprintf(`<link rel="stylesheet" href="/%s">`, css))
					sb.WriteByte('\n')
				}
			}

			// 2. 添加模块预加载（在主脚本之前，用于性能优化）
			visitedModules := make(map[string]bool)
			allModulePreloads := collectAllModulePreloads(origin, visitedModules)
			moduleSet := make(map[string]bool)
			for _, moduleFile := range allModulePreloads {
				if !moduleSet[moduleFile] && moduleFile != item.File {
					moduleSet[moduleFile] = true
					sb.WriteString(fmt.Sprintf(`<link rel="modulepreload" href="/%s">`, moduleFile))
					sb.WriteByte('\n')
				}
			}

			// 3. 添加主入口文件
			fileExt := filepath.Ext(item.File)
			switch fileExt {
			case ".js", ".mjs", ".ts", ".jsx", ".tsx":
				sb.WriteString(fmt.Sprintf(`<script type="module" src="/%s"></script>`, item.File))
			case ".css":
				// CSS 文件已在上面处理，这里不重复添加
				if !cssSet[item.File] {
					sb.WriteString(fmt.Sprintf(`<link rel="stylesheet" href="/%s">`, item.File))
				}
			default:
				// 非标准文件类型使用原处理逻辑
				sb.WriteString(fmt.Sprintf(`<script type="module" src="/%s"></script>`, item.File))
			}
			sb.WriteByte('\n')

			// 4. 预加载静态资源（字体等，可选）
			visitedAssets := make(map[string]bool)
			allAssets := collectAllAssets(origin, visitedAssets)
			assetSet := make(map[string]bool)
			for _, assetFile := range allAssets {
				if !assetSet[assetFile] {
					assetSet[assetFile] = true
					// 根据文件类型添加适当的预加载
					assetExt := filepath.Ext(assetFile)
					switch assetExt {
					case ".woff", ".woff2":
						// 字体文件预加载
						sb.WriteString(fmt.Sprintf(`<link rel="preload" href="/%s" as="font" type="font/%s" crossorigin>`, assetFile, assetExt[1:]))
						sb.WriteByte('\n')
					case ".jpg", ".jpeg", ".png", ".webp", ".avif":
						// 图片文件可以选择性预加载（通常不建议自动预加载所有图片）
						// sb.WriteString(fmt.Sprintf(`<link rel="preload" href="/%s" as="image">`, assetFile))
						// sb.WriteByte('\n')
					default:
						// 其他资源类型暂不处理
					}
				}
			}
		} else {
			// 如果在 manifest 中找不到，可能是开发时新增的文件，回退到直接引用
			fileExt := filepath.Ext(origin)
			switch fileExt {
			case ".js", ".mjs", ".ts", ".jsx", ".tsx":
				sb.WriteString(fmt.Sprintf(`<script type="module" src="/%s"></script>`, origin))
			case ".css":
				sb.WriteString(fmt.Sprintf(`<link rel="stylesheet" href="/%s">`, origin))
			default:
				sb.WriteString(fmt.Sprintf(`<script type="module" src="/%s"></script>`, origin))
			}
			sb.WriteByte('\n')
		}
	}

	res := template.HTML(sb.String())
	htmlHeaderCache.Store(cacheKey, res)
	return res
}
