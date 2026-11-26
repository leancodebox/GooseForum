package forum

import (
	"fmt"
	"html/template"
	"io/fs"
	"maps"
	"path/filepath"
	"strings"

	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/resource"
)

type manifestItem struct {
	File    string   `json:"file"`
	IsEntry bool     `json:"isEntry"`
	Css     []string `json:"css"`
	Assets  []string `json:"assets"`
	Imports []string `json:"imports"`
}

// React assets and their manifest are embedded in production; rebuild Go after the frontend build.
var reactManifest = loadManifestAt("static/dist/react/.vite/manifest.json")

func resourceEntry(layout string) template.HTML {
	const entry, source = "index.html", "src/main.tsx"
	if setting.IsProduction() {
		return manifestEntry(reactManifest, entry, layout, "react/")
	}
	server := strings.TrimRight(preferences.GetString("resource.reactDevServer", "http://localhost:3011"), "/")
	return template.HTML(fmt.Sprintf(`<script type="module">
import RefreshRuntime from "%s/@react-refresh";
RefreshRuntime.injectIntoGlobalHook(window);
window.$RefreshReg$ = () => {};
window.$RefreshSig$ = () => (type) => type;
window.__vite_plugin_react_preamble_installed__ = true;
</script>
<script type="module" src="%s/@vite/client"></script>
<script type="module" src="%s/%s"></script>`, server, server, server, source))
}

// Vite records each layout's CSS URL in the lazy entry's assets.
func layoutStylesheet(entry string) string {
	if setting.IsProduction() {
		for _, asset := range reactManifest[entry].Assets {
			if strings.HasSuffix(asset, ".css") {
				return resourceAsset("react/" + asset)
			}
		}
	}
	return ""
}

func manifestEntry(entries map[string]manifestItem, origin, layout, prefix string) template.HTML {
	item, ok := entries[origin]
	if !ok {
		return template.HTML(fmt.Sprintf(`<script type="module" src="%s"></script>`, resourceAsset(origin)))
	}

	var sb strings.Builder
	for _, css := range collectManifestCSS(entries, origin, map[string]bool{}) {
		fmt.Fprintf(&sb, `<link rel="stylesheet" href="%s" crossorigin>`, resourceAsset(prefix+css))
		sb.WriteByte('\n')
	}
	// The server already knows which lazy layout the bootstrap will import.
	if chunk := entries[layout]; chunk.File != "" {
		fmt.Fprintf(&sb, `<link rel="modulepreload" href="%s" crossorigin>`, resourceAsset(prefix+chunk.File))
		sb.WriteByte('\n')
	}
	fmt.Fprintf(&sb, `<script type="module" src="%s" crossorigin></script>`, resourceAsset(prefix+item.File))
	sb.WriteByte('\n')
	return template.HTML(sb.String())
}

func resourceAsset(path string) string {
	if strings.HasPrefix(path, "/") {
		return path
	}
	assetPath := "/assets/" + strings.TrimPrefix(path, "/")
	if cdnURL := strings.TrimRight(strings.TrimSpace(setting.GetCDNURL()), "/"); cdnURL != "" {
		return cdnURL + assetPath
	}
	return assetPath
}

func loadManifestAt(path string) map[string]manifestItem {
	content, err := fs.ReadFile(resource.GetTemplateFS(), path)
	if err != nil {
		return map[string]manifestItem{}
	}
	return jsonopt.Decode[map[string]manifestItem](content)
}

func collectManifestCSS(entries map[string]manifestItem, entry string, visited map[string]bool) []string {
	if visited[entry] {
		return nil
	}
	visited[entry] = true
	item, ok := entries[entry]
	if !ok {
		return nil
	}

	files := append([]string{}, item.Css...)
	for _, importKey := range item.Imports {
		files = append(files, collectManifestCSS(entries, importKey, visited)...)
	}
	return dedupeStrings(files)
}

func dedupeStrings(values []string) []string {
	seen := map[string]bool{}
	res := make([]string, 0, len(values))
	for _, value := range values {
		normalized := filepath.ToSlash(value)
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		res = append(res, normalized)
	}
	return res
}

func templateFuncs() template.FuncMap {
	funcs := template.FuncMap{}
	maps.Copy(funcs, templateFuncMap)
	funcs["ResourceEntry"] = resourceEntry
	funcs["LayoutStylesheet"] = layoutStylesheet
	funcs["ResourceAsset"] = resourceAsset
	return funcs
}
