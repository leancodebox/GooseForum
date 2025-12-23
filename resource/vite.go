package resource

import (
	_ "embed"
	"fmt"
	"html/template"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"

	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
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

// ThemeConfig represents the theme configuration
type ThemeConfig struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Manifest    string `json:"manifest"`
}

var (
	manifestItemMap = map[string]ManifestItem{}
	htmlHeaderCache sync.Map
	// ViteDevServerURL is the URL of the Vite dev server
	ViteDevServerURL = "http://localhost:3009"
)

type Theme struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	Version          string `json:"version"`
	ViteDevServerURL string `json:"ViteDevServerURL"`
	Manifest         string `json:"manifest"`
}

func init() {
	// Try to load manifest if available
	manifestPath := "static/dist/.vite/manifest.json"

	// Parse theme.json if available
	if len(themeConfig) > 0 {
		theme := jsonopt.Decode[ThemeConfig](themeConfig)
		if theme.Manifest != "" {
			manifestPath = theme.Manifest
		}
	}

	content, err := viewAssert.ReadFile(manifestPath)
	if err != nil {
		// Only log error in production if we expect it to exist
		if setting.IsProduction() {
			slog.Error("ManifestGetError", "error", err, "path", manifestPath)
		}
		return
	}
	info := jsonopt.Decode[map[string]ManifestItem](content)
	manifestItemMap = info

	// Prebuild cache in production
	if setting.IsProduction() {
		prebuildProductionCache()
	}
}

// prebuildProductionCache builds HTML cache for all entries in production
func prebuildProductionCache() {
	for key, item := range manifestItemMap {
		if item.IsEntry {
			cacheKey := fmt.Sprintf("%s_%v", key, true)
			html := generateProductionHTML(key)
			htmlHeaderCache.Store(cacheKey, template.HTML(html))
		}
	}
	slog.Info("Production HTML cache prebuilt successfully", "cached_items", getCacheSize())
}

func getCacheSize() int {
	count := 0
	htmlHeaderCache.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

func collectAllCSSFiles(entryKey string, visited map[string]bool) []string {
	if visited[entryKey] {
		return nil
	}
	visited[entryKey] = true

	var cssFiles []string
	if item, ok := manifestItemMap[entryKey]; ok {
		cssFiles = append(cssFiles, item.Css...)
		for _, importKey := range item.Imports {
			cssFiles = append(cssFiles, collectAllCSSFiles(importKey, visited)...)
		}
		for _, dynamicImportKey := range item.DynamicImports {
			cssFiles = append(cssFiles, collectAllCSSFiles(dynamicImportKey, visited)...)
		}
	}
	return cssFiles
}

func collectAllModulePreloads(entryKey string, visited map[string]bool) []string {
	if visited[entryKey] {
		return nil
	}
	visited[entryKey] = true

	var moduleFiles []string
	if item, ok := manifestItemMap[entryKey]; ok {
		for _, importKey := range item.Imports {
			if importedItem, exists := manifestItemMap[importKey]; exists {
				moduleFiles = append(moduleFiles, importedItem.File)
				moduleFiles = append(moduleFiles, collectAllModulePreloads(importKey, visited)...)
			}
		}
	}
	return moduleFiles
}

func collectAllAssets(entryKey string, visited map[string]bool) []string {
	if visited[entryKey] {
		return nil
	}
	visited[entryKey] = true

	var assetFiles []string
	if item, ok := manifestItemMap[entryKey]; ok {
		assetFiles = append(assetFiles, item.Assets...)
		for _, importKey := range item.Imports {
			assetFiles = append(assetFiles, collectAllAssets(importKey, visited)...)
		}
		for _, dynamicImportKey := range item.DynamicImports {
			assetFiles = append(assetFiles, collectAllAssets(dynamicImportKey, visited)...)
		}
	}
	return assetFiles
}

// ViteEntry generates HTML tags for an entry point
func ViteEntry(origin string) template.HTML {
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

// VitePath returns the resolved URL for an asset
func VitePath(path string) string {
	if setting.IsProduction() {
		if item, ok := manifestItemMap[path]; ok {
			return Asset(item.File)
		}
		// If not found in manifest, assume it's a direct asset
		return Asset(path)
	}
	// Development: proxy to vite server
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(ViteDevServerURL, "/"), strings.TrimPrefix(path, "/"))
}

func generateDevelopmentHTML(origin string) string {
	return generateFileTag(origin, ViteDevServerURL)
}

func generateProductionHTML(origin string) string {
	item, exists := manifestItemMap[origin]
	if !exists {
		return generateFileTag(origin, "")
	}
	return buildResourceTags(origin, item)
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

func buildResourceTags(origin string, item ManifestItem) string {
	sb := &strings.Builder{}

	visitedCSS := make(map[string]bool)
	visitedModules := make(map[string]bool)
	visitedAssets := make(map[string]bool)

	cssFiles := collectAllCSSFiles(origin, visitedCSS)
	moduleFiles := collectAllModulePreloads(origin, visitedModules)
	assets := collectAllAssets(origin, visitedAssets)

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
