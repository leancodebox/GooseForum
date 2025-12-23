package viewrender

import (
	_ "embed"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"

	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/resource"
)

// ManifestItem represents an entry in the manifest.json
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

type ViteHandler struct {
	manifestItemMap  map[string]ManifestItem
	htmlHeaderCache  sync.Map
	viteDevServerURL string
	isProduction     bool
}

var (
	// DefaultViteHandler is the default instance of ViteHandler
	DefaultViteHandler *ViteHandler
)

func init() {
	// Try to load manifest if available
	manifestPath := "static/dist/.vite/manifest.json"

	// Parse theme.json if available
	theme, _ := resource.GetThemeConfig()
	if theme != nil && theme.Manifest != "" {
		manifestPath = theme.Manifest
	}

	isProduction := setting.IsProduction()
	var manifestItemMap map[string]ManifestItem

	content, err := fs.ReadFile(resource.GetTemplateFS(), manifestPath)
	if err != nil {
		// Only log error in production if we expect it to exist
		if isProduction {
			slog.Error("ManifestGetError", "error", err, "path", manifestPath)
		}
	} else {
		manifestItemMap = jsonopt.Decode[map[string]ManifestItem](content)
	}

	DefaultViteHandler = NewViteHandler(
		"http://localhost:3009",
		isProduction,
		manifestItemMap,
	)
}

// NewViteHandler creates a new ViteHandler
func NewViteHandler(viteDevServerURL string, isProduction bool, manifestItemMap map[string]ManifestItem) *ViteHandler {
	if manifestItemMap == nil {
		manifestItemMap = make(map[string]ManifestItem)
	}
	handler := &ViteHandler{
		manifestItemMap:  manifestItemMap,
		viteDevServerURL: viteDevServerURL,
		isProduction:     isProduction,
	}

	// Prebuild cache in production
	if isProduction {
		handler.prebuildProductionCache()
	}

	return handler
}

// prebuildProductionCache builds HTML cache for all entries in production
func (v *ViteHandler) prebuildProductionCache() {
	for key, item := range v.manifestItemMap {
		if item.IsEntry {
			cacheKey := fmt.Sprintf("%s_%v", key, true)
			html := v.generateProductionHTML(key)
			v.htmlHeaderCache.Store(cacheKey, template.HTML(html))
		}
	}
	slog.Info("Production HTML cache prebuilt successfully", "cached_items", v.getCacheSize())
}

func (v *ViteHandler) getCacheSize() int {
	count := 0
	v.htmlHeaderCache.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

func (v *ViteHandler) collectAllCSSFiles(entryKey string, visited map[string]bool) []string {
	if visited[entryKey] {
		return nil
	}
	visited[entryKey] = true

	var cssFiles []string
	if item, ok := v.manifestItemMap[entryKey]; ok {
		cssFiles = append(cssFiles, item.Css...)
		for _, importKey := range item.Imports {
			cssFiles = append(cssFiles, v.collectAllCSSFiles(importKey, visited)...)
		}
		for _, dynamicImportKey := range item.DynamicImports {
			cssFiles = append(cssFiles, v.collectAllCSSFiles(dynamicImportKey, visited)...)
		}
	}
	return cssFiles
}

func (v *ViteHandler) collectAllModulePreloads(entryKey string, visited map[string]bool) []string {
	if visited[entryKey] {
		return nil
	}
	visited[entryKey] = true

	var moduleFiles []string
	if item, ok := v.manifestItemMap[entryKey]; ok {
		for _, importKey := range item.Imports {
			if importedItem, exists := v.manifestItemMap[importKey]; exists {
				moduleFiles = append(moduleFiles, importedItem.File)
				moduleFiles = append(moduleFiles, v.collectAllModulePreloads(importKey, visited)...)
			}
		}
	}
	return moduleFiles
}

func (v *ViteHandler) collectAllAssets(entryKey string, visited map[string]bool) []string {
	if visited[entryKey] {
		return nil
	}
	visited[entryKey] = true

	var assetFiles []string
	if item, ok := v.manifestItemMap[entryKey]; ok {
		assetFiles = append(assetFiles, item.Assets...)
		for _, importKey := range item.Imports {
			assetFiles = append(assetFiles, v.collectAllAssets(importKey, visited)...)
		}
		for _, dynamicImportKey := range item.DynamicImports {
			assetFiles = append(assetFiles, v.collectAllAssets(dynamicImportKey, visited)...)
		}
	}
	return assetFiles
}

// ViteEntry generates HTML tags for an entry point
func ViteEntry(origin string) template.HTML {
	return DefaultViteHandler.ViteEntry(origin)
}

// ViteEntry generates HTML tags for an entry point
func (v *ViteHandler) ViteEntry(origin string) template.HTML {
	cacheKey := fmt.Sprintf("%s_%v", origin, v.isProduction)
	if val, cached := v.htmlHeaderCache.Load(cacheKey); cached {
		return val.(template.HTML)
	}

	var html string
	if v.isProduction {
		html = v.generateProductionHTML(origin)
	} else {
		html = v.generateDevelopmentHTML(origin)
	}

	res := template.HTML(html)
	v.htmlHeaderCache.Store(cacheKey, res)
	return res
}

// VitePath returns the resolved URL for an asset
func VitePath(path string) string {
	return DefaultViteHandler.VitePath(path)
}

// VitePath returns the resolved URL for an asset
func (v *ViteHandler) VitePath(path string) string {
	if v.isProduction {
		if item, ok := v.manifestItemMap[path]; ok {
			return Asset(item.File)
		}
		// If not found in manifest, assume it's a direct asset
		return Asset(path)
	}
	// Development: proxy to vite server
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(v.viteDevServerURL, "/"), strings.TrimPrefix(path, "/"))
}

func (v *ViteHandler) generateDevelopmentHTML(origin string) string {
	return generateFileTag(origin, v.viteDevServerURL)
}

func (v *ViteHandler) generateProductionHTML(origin string) string {
	item, exists := v.manifestItemMap[origin]
	if !exists {
		return generateFileTag(origin, "")
	}
	return v.buildResourceTags(origin, item)
}

func generateFileTag(filename, baseURL string) string {
	var url string
	if baseURL != "" {
		url = fmt.Sprintf("%s/%s", strings.TrimSuffix(baseURL, "/"), strings.TrimPrefix(filename, "/"))
	} else {
		url = Asset(filename)
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

func (v *ViteHandler) buildResourceTags(origin string, item ManifestItem) string {
	sb := &strings.Builder{}

	visitedCSS := make(map[string]bool)
	visitedModules := make(map[string]bool)
	visitedAssets := make(map[string]bool)

	cssFiles := v.collectAllCSSFiles(origin, visitedCSS)
	moduleFiles := v.collectAllModulePreloads(origin, visitedModules)
	assets := v.collectAllAssets(origin, visitedAssets)

	cssSet := make(map[string]bool)
	for _, css := range cssFiles {
		if !cssSet[css] {
			cssSet[css] = true
			fmt.Fprintf(sb, `<link rel="stylesheet" href="%s" crossorigin>`, Asset(css))
			sb.WriteByte('\n')
		}
	}

	moduleSet := make(map[string]bool)
	for _, moduleFile := range moduleFiles {
		if !moduleSet[moduleFile] && moduleFile != item.File {
			moduleSet[moduleFile] = true
			fmt.Fprintf(sb, `<link rel="modulepreload" href="%s" crossorigin>`, Asset(moduleFile))
			sb.WriteByte('\n')
		}
	}

	assetSet := make(map[string]bool)
	for _, assetFile := range assets {
		if !assetSet[assetFile] {
			assetSet[assetFile] = true
			addAssetPreloadTag(sb, assetFile)
		}
	}

	switch filepath.Ext(item.File) {
	case ".js", ".mjs", ".ts", ".jsx", ".tsx":
		fmt.Fprintf(sb, `<script type="module" src="%s" crossorigin></script>`, Asset(item.File))
		sb.WriteByte('\n')
	case ".css":
	default:
		fmt.Fprintf(sb, `<script type="module" src="%s" crossorigin></script>`, Asset(item.File))
		sb.WriteByte('\n')
	}

	return sb.String()
}

func addAssetPreloadTag(sb *strings.Builder, assetFile string) {
	ext := filepath.Ext(assetFile)
	switch ext {
	case ".woff", ".woff2":
		fmt.Fprintf(sb, `<link rel="preload" href="%s" as="font" type="font/%s" crossorigin>`, Asset(assetFile), ext[1:])
		sb.WriteByte('\n')
	case ".ttf", ".otf":
		fmt.Fprintf(sb, `<link rel="preload" href="%s" as="font" crossorigin>`, Asset(assetFile))
		sb.WriteByte('\n')
	case ".jpg", ".jpeg", ".png", ".webp", ".avif", ".svg":
		fmt.Fprintf(sb, `<link rel="preload" href="%s" as="image">`, Asset(assetFile))
		sb.WriteByte('\n')
	case ".css":
		fmt.Fprintf(sb, `<link rel="preload" href="%s" as="style" crossorigin>`, Asset(assetFile))
		sb.WriteByte('\n')
	case ".js", ".mjs":
		fmt.Fprintf(sb, `<link rel="preload" href="%s" as="script" crossorigin>`, Asset(assetFile))
		sb.WriteByte('\n')
	}
}
