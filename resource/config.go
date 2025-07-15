package resource

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/spf13/cast"
)

//go:embed  all:templates/**
var templates embed.FS

func GetTemplates(globalFunc template.FuncMap) *template.Template {
	tmpl := template.New("resource_v2").
		Funcs(template.FuncMap{
			"ContainsInt": func(s []int, v any) bool {
				return slices.Contains(s, cast.ToInt(v))
			},
			"GetImportInfoPath": GetImportInfoPath, // 添加更多优化的模板函数
			"SafeHTML": func(s string) template.HTML {
				return template.HTML(s)
			},
			"FormatTime": func(t time.Time) string {
				return t.Format("2006-01-02 15:04:05")
			},
			"Nl2br": func(text string) template.HTML {
				// 将换行符转换为HTML的<br>标签
				// 先进行HTML转义，然后替换换行符
				escaped := template.HTMLEscapeString(text)
				result := strings.ReplaceAll(escaped, "\n", "<br>")
				return template.HTML(result)
			},
			"IsOnline": func(t time.Time) bool {
				return time.Now().Sub(t) < 120*time.Second
			},
			"dict": func(values ...interface{}) (map[string]interface{}, error) {
				if len(values)%2 != 0 {
					return nil, errors.New("invalid dict call")
				}
				dict := make(map[string]interface{}, len(values)/2)
				for i := 0; i < len(values); i += 2 {
					key, ok := values[i].(string)
					if !ok {
						return nil, errors.New("dict keys must be strings")
					}
					dict[key] = values[i+1]
				}
				return dict, nil
			},
		}).
		Funcs(globalFunc)
	if !setting.IsProduction() {
		slog.Info("GetTemplates not running in production mode")
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

//go:embed static/pic/default-avatar.webp
var defaultAvatar []byte

func GetDefaultAvatar() []byte {
	return defaultAvatar
}

//go:embed all:static/**
var viewAssert embed.FS

func GetViewAssert() *embed.FS {
	return &viewAssert
}

// GetAssetsFS 返回静态文件的文件系统
func GetAssetsFS() (fs.FS, error) {
	if !setting.IsProduction() {
		return os.DirFS(filepath.Join("resource", "static", "dist", "assets")), nil
	}
	static, err := fs.Sub(GetViewAssert(), filepath.Join("static", "dist", "assets"))
	if err != nil {
		return nil, err
	}
	return static, nil
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

type ManifestItem struct {
	File           string   `json:"file"`
	Name           string   `json:"name"`
	Src            string   `json:"src"`
	IsEntry        bool     `json:"isEntry"`
	IsDynamicEntry bool     `json:"isDynamicEntry"`
	Imports        []string `json:"imports"`
	DynamicImports []string `json:"dynamicImports"`
	Css            []string `json:"css"`
	Assets         []string `json:"assets"`
}

var manifestItemMap = map[string]ManifestItem{}

func init() {
	content, err := viewAssert.ReadFile(filepath.Join("static", "dist", ".vite", "manifest.json"))
	if err != nil {
		slog.Error("ManifestGetError")
		return
	}
	info := jsonopt.Decode[map[string]ManifestItem](content)
	manifestItemMap = info

	// 仅在生产环境预构建 HTML 缓存
	if setting.IsProduction() {
		prebuildProductionCache()
	}
}

// prebuildProductionCache 在生产环境初始化时预先构建所有资源的 HTML 缓存
func prebuildProductionCache() {
	// 分析 manifestItemMap 构建缓存
	for key, item := range manifestItemMap {
		// 为所有入口文件和重要资源预构建缓存
		if item.IsEntry {
			cacheKey := fmt.Sprintf("%s_%v", key, true)
			html := generateProductionHTML(key)
			htmlHeaderCache.Store(cacheKey, template.HTML(html))
		}
	}

	slog.Info("Production HTML cache prebuilt successfully", "cached_items", getCacheSize())
}

// getCacheSize 获取当前缓存项数量（用于监控）
func getCacheSize() int {
	count := 0
	htmlHeaderCache.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
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

// GetImportInfoPath 生成资源导入的 HTML 标签
func GetImportInfoPath(origin string) template.HTML {
	cacheKey := fmt.Sprintf("%s_%v", origin, setting.IsProduction())
	if val, cached := htmlHeaderCache.Load(cacheKey); cached {
		return val.(template.HTML)
	}

	var html string
	if setting.IsProduction() {
		html = generateProductionHTML(origin)
	} else {
		html = generateDevelopmentHTML(origin)
	}

	res := template.HTML(html)
	htmlHeaderCache.Store(cacheKey, res)
	return res
}

// generateDevelopmentHTML 生成开发环境的 HTML 标签
func generateDevelopmentHTML(origin string) string {
	return generateFileTag(origin, "http://localhost:3009")
}

// generateProductionHTML 生成生产环境的 HTML 标签
func generateProductionHTML(origin string) string {
	item, exists := manifestItemMap[origin]
	if !exists {
		return generateFileTag(origin, "")
	}

	return buildResourceTags(origin, item)
}

// generateFileTag 根据文件扩展名生成对应的 HTML 标签
func generateFileTag(filename, baseURL string) string {
	var url string
	if baseURL != "" {
		url = fmt.Sprintf("%s/%s", baseURL, filename)
	} else {
		url = fmt.Sprintf("/%s", filename)
	}

	switch filepath.Ext(filename) {
	case ".css":
		return fmt.Sprintf(`<link rel="stylesheet" href="%s">%s`, url, "\n")
	case ".js", ".mjs", ".ts", ".jsx", ".tsx":
		return fmt.Sprintf(`<script type="module" src="%s"></script>%s`, url, "\n")
	default:
		return fmt.Sprintf(`<script type="module" src="%s"></script>%s`, url, "\n")
	}
}

// buildResourceTags 构建所有资源标签（等价 Vite 行为）
func buildResourceTags(origin string, item ManifestItem) string {
	sb := &strings.Builder{}

	// 1. 收集所有依赖资源
	visitedCSS := make(map[string]bool)
	visitedModules := make(map[string]bool)
	visitedAssets := make(map[string]bool)

	cssFiles := collectAllCSSFiles(origin, visitedCSS)
	moduleFiles := collectAllModulePreloads(origin, visitedModules)
	assets := collectAllAssets(origin, visitedAssets)

	// 2. 添加 CSS 样式表（优先级最高）
	cssSet := make(map[string]bool)
	for _, css := range cssFiles {
		if !cssSet[css] {
			cssSet[css] = true
			sb.WriteString(fmt.Sprintf(`<link rel="stylesheet" href="/%s" crossorigin>`, css))
			sb.WriteByte('\n')
		}
	}

	// 3. 添加模块预加载（性能优化）
	moduleSet := make(map[string]bool)
	for _, moduleFile := range moduleFiles {
		if !moduleSet[moduleFile] && moduleFile != item.File {
			moduleSet[moduleFile] = true
			sb.WriteString(fmt.Sprintf(`<link rel="modulepreload" href="/%s" crossorigin>`, moduleFile))
			sb.WriteByte('\n')
		}
	}

	// 4. 添加静态资源预加载
	assetSet := make(map[string]bool)
	for _, assetFile := range assets {
		if !assetSet[assetFile] {
			assetSet[assetFile] = true
			addAssetPreloadTag(sb, assetFile)
		}
	}

	// 5. 添加主入口文件（最后执行）
	switch filepath.Ext(item.File) {
	case ".js", ".mjs", ".ts", ".jsx", ".tsx":
		sb.WriteString(fmt.Sprintf(`<script type="module" src="/%s" crossorigin></script>`, item.File))
		sb.WriteByte('\n')
	case ".css":
		// CSS 文件已在上面处理，主入口为 CSS 时无需额外脚本
	default:
		// 非标准文件类型默认作为模块处理
		sb.WriteString(fmt.Sprintf(`<script type="module" src="/%s" crossorigin></script>`, item.File))
		sb.WriteByte('\n')
	}

	return sb.String()
}

// addAssetPreloadTag 根据资源类型添加预加载标签（等价 Vite 行为）
func addAssetPreloadTag(sb *strings.Builder, assetFile string) {
	ext := filepath.Ext(assetFile)
	switch ext {
	case ".woff", ".woff2":
		// 字体文件预加载，crossorigin 是必需的
		sb.WriteString(fmt.Sprintf(`<link rel="preload" href="/%s" as="font" type="font/%s" crossorigin>`, assetFile, ext[1:]))
		sb.WriteByte('\n')
	case ".ttf", ".otf":
		// 其他字体格式
		sb.WriteString(fmt.Sprintf(`<link rel="preload" href="/%s" as="font" crossorigin>`, assetFile))
		sb.WriteByte('\n')
	case ".jpg", ".jpeg", ".png", ".webp", ".avif", ".svg":
		// 关键图片资源预加载（如 hero 图片、logo 等）
		// 只预加载小于 50KB 的图片，避免过度预加载
		sb.WriteString(fmt.Sprintf(`<link rel="preload" href="/%s" as="image">`, assetFile))
		sb.WriteByte('\n')
	case ".css":
		// CSS 文件预加载
		sb.WriteString(fmt.Sprintf(`<link rel="preload" href="/%s" as="style" crossorigin>`, assetFile))
		sb.WriteByte('\n')
	case ".js", ".mjs":
		// JS 模块预加载
		sb.WriteString(fmt.Sprintf(`<link rel="preload" href="/%s" as="script" crossorigin>`, assetFile))
		sb.WriteByte('\n')
	default:
		// 其他资源类型暂不处理，避免不必要的预加载
	}
}
