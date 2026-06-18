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
	Imports []string `json:"imports"`
}

var manifest = loadManifest()

func resourceEntry(origin string) template.HTML {
	if devServer := viteDevServer(); devServer != "" {
		devServer = strings.TrimRight(devServer, "/")
		origin = strings.TrimPrefix(origin, "/")
		devBase := strings.Trim(viteDevBase(), "/")
		if devBase != "" {
			devBase += "/"
		}
		return template.HTML(fmt.Sprintf(`<script type="module" src="%s/%s@vite/client"></script>
<script type="module" src="%s/%s%s"></script>`, devServer, devBase, devServer, devBase, origin))
	}

	item, ok := manifest[origin]
	if !ok {
		return template.HTML(fmt.Sprintf(`<script type="module" src="%s"></script>`, resourceAsset(origin)))
	}

	var sb strings.Builder
	for _, css := range collectCSS(origin, map[string]bool{}) {
		fmt.Fprintf(&sb, `<link rel="stylesheet" href="%s" crossorigin>`, resourceAsset(css))
		sb.WriteByte('\n')
	}
	fmt.Fprintf(&sb, `<script type="module" src="%s" crossorigin></script>`, resourceAsset(item.File))
	sb.WriteByte('\n')
	return template.HTML(sb.String())
}

func viteDevServer() string {
	return viteDevServerFor(preferences.GetString("resource.devServer", ""), setting.IsProduction())
}

func viteDevServerFor(devServer string, production bool) string {
	if devServer != "" {
		return devServer
	}
	if !production {
		return "http://localhost:3010"
	}
	return ""
}

func viteDevBase() string {
	return preferences.GetString("resource.devBase", "/assets/")
}

func resourceAsset(path string) string {
	if strings.HasPrefix(path, "/") {
		return path
	}
	return "/assets/" + strings.TrimPrefix(path, "/")
}

func loadManifest() map[string]manifestItem {
	content, err := fs.ReadFile(resource.GetTemplateFS(), "static/dist/.vite/manifest.json")
	if err != nil {
		return map[string]manifestItem{}
	}
	return jsonopt.Decode[map[string]manifestItem](content)
}

func collectCSS(entry string, visited map[string]bool) []string {
	if visited[entry] {
		return nil
	}
	visited[entry] = true
	item, ok := manifest[entry]
	if !ok {
		return nil
	}

	files := append([]string{}, item.Css...)
	for _, importKey := range item.Imports {
		files = append(files, collectCSS(importKey, visited)...)
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
	funcs["ResourceAsset"] = resourceAsset
	return funcs
}
