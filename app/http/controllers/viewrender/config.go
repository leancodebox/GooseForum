package viewrender

import (
	"html/template"
	"slices"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"
	"github.com/spf13/cast"
)

// TemplateFuncs defines the available functions for templates
var TemplateFuncs = template.FuncMap{
	"Url": func() URLHelper {
		return URLHelper{}
	},
	"ContainsInt": func(s []int, v any) bool {
		return slices.Contains(s, cast.ToInt(v))
	},
	"ViteEntry": ViteEntry,
	"VitePath":  VitePath,
	"Asset":     Asset,
	"SafeHTML": func(s string) template.HTML {
		return template.HTML(s)
	},
	"FormatTime": func(t time.Time) string {
		return t.Format(time.DateTime)
	},
	"Nl2br": func(text string) template.HTML {
		// 将换行符转换为HTML的<br>标签
		// 先进行HTML转义，然后替换换行符
		escaped := template.HTMLEscapeString(text)
		result := strings.ReplaceAll(escaped, "\n", "<br>")
		return template.HTML(result)
	},
	"IsOnline": func(t time.Time) bool {
		return time.Since(t) < 120*time.Second
	},
	"add": func(a, b int) int {
		return a + b
	},
	"sub": func(a, b int) int {
		return a - b
	},
	"ToJson": func(v any) string {
		return jsonopt.EncodeFormat(v)
	},
	"json": func(v any) template.JS {
		return template.JS(jsonopt.Encode(v))
	},
}

// Asset 返回静态资源的完整路径，如果配置了 CDN 则返回 CDN 路径
func Asset(path string) string {
	if path == "" {
		return ""
	}

	// 如果 path 已经是一个完整的 URL，直接返回
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "//") {
		return path
	}

	cdnURL := setting.GetCDNURL()
	if cdnURL == "" {
		if strings.HasPrefix(path, "/") {
			return path
		}
		return "/" + path
	}

	// 移除 path 开头的斜杠，避免拼接出两个斜杠
	path = strings.TrimPrefix(path, "/")
	// 移除 cdnURL 结尾的斜杠
	cdnURL = strings.TrimSuffix(cdnURL, "/")

	return cdnURL + "/" + path
}

type URLHelper struct{}

func (u URLHelper) Home() string                   { return urlconfig.Home() }
func (u URLHelper) Post() string                   { return urlconfig.Post() }
func (u URLHelper) Docs() string                   { return urlconfig.Docs() }
func (u URLHelper) Links() string                  { return urlconfig.Links() }
func (u URLHelper) Sponsors() string               { return urlconfig.Sponsors() }
func (u URLHelper) About() string                  { return urlconfig.About() }
func (u URLHelper) Publish() string                { return urlconfig.Publish() }
func (u URLHelper) Search() string                 { return urlconfig.Search() }
func (u URLHelper) Register() string               { return urlconfig.Register() }
func (u URLHelper) Login() string                  { return urlconfig.Login() }
func (u URLHelper) Admin() string                  { return urlconfig.Admin() }
func (u URLHelper) Messages() string               { return urlconfig.Messages() }
func (u URLHelper) Settings() string               { return urlconfig.Settings() }
func (u URLHelper) Notifications() string          { return urlconfig.Notifications() }
func (u URLHelper) Activate() string               { return urlconfig.Activate() }
func (u URLHelper) ResetPassword() string          { return urlconfig.ResetPassword() }
func (u URLHelper) PostDetail(id any) string       { return urlconfig.PostDetail(id) }
func (u URLHelper) User(id any) string             { return urlconfig.User(id) }
func (u URLHelper) DocsProject(slug any) string    { return urlconfig.DocsProject(slug) }
func (u URLHelper) DocsContent(p, v, c any) string { return urlconfig.DocsContent(p, v, c) }
func (u URLHelper) Rss() string                    { return urlconfig.Rss() }
