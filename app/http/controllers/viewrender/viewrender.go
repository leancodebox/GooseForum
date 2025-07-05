package viewrender

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/datacache"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/resource"
	"html/template"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

var ht4gooseforum *template.Template
var htht4gooseforumOnce sync.Once

func Reload() {
	ht4gooseforum = resource.GetTemplates(GlobalFunc())
}

func init() {
	htht4gooseforumOnce.Do(func() {
		// 创建基础模板
		Reload()
	})
}

var webSettingsCache = &datacache.Cache[string, pageConfig.WebSettingsConfig]{}

func GlobalFunc() template.FuncMap {
	return template.FuncMap{
		"WebPageSettings": func() pageConfig.WebSettingsConfig {
			data, _ := webSettingsCache.GetOrLoad("websetcache", func() (pageConfig.WebSettingsConfig, error) {
				return pageConfig.GetConfigByPageType(pageConfig.WebSettings, pageConfig.WebSettingsConfig{
					MetaTags:      "",
					CustomCSS:     "",
					CustomJS:      "",
					ExternalLinks: "",
					Favicon:       "",
				}), nil
			},
				time.Second*5,
			)
			return data
		},
	}
}

func Render(c *gin.Context, name string, data any) {
	// 从cookie中读取主题设置
	theme := "light" // 默认主题
	if themeCookie, err := c.Cookie("theme"); err == nil && themeCookie != "" {
		theme = themeCookie
	}
	// 将数据转换为map并添加主题信息
	var templateData map[string]any
	if dataMap, ok := data.(map[string]any); ok {
		templateData = dataMap
	} else {
		templateData = map[string]any{"Data": data}
	}
	templateData["Theme"] = theme
	if err := ht4gooseforum.ExecuteTemplate(c.Writer, name, templateData); err != nil {
		slog.Error("render template err", "err", err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}
