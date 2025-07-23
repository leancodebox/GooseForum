package viewrender

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/http/controllers/datacache"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/resource"
	"html/template"
	"log/slog"
	"net/http"
	"time"
)

var ht4gooseforum *template.Template

func Reload() {
	ht4gooseforum = resource.GetTemplates(GlobalFunc())
}

func init() {
	Reload()
}

var webSettingsCache = &datacache.Cache[pageConfig.WebSettingsConfig]{}

func GlobalFunc() template.FuncMap {
	return template.FuncMap{
		"WebPageSettings": func() pageConfig.WebSettingsConfig {
			data, _ := webSettingsCache.GetOrLoadE("websetcache", func() (pageConfig.WebSettingsConfig, error) {
				return pageConfig.GetConfigByPageType(pageConfig.WebSettings, pageConfig.WebSettingsConfig{}), nil
			},
				time.Second*5,
			)
			return data
		},
	}
}

func Render(c *gin.Context, name string, templateData map[string]any) {
	if templateData == nil {
		templateData = make(map[string]any, 4)
	}
	templateData["IsProduction"] = setting.IsProduction()
	templateData["Theme"] = GetTheme(c)
	templateData["Footer"] = hotdataserve.GetFooterConfigCache()
	templateData["SiteSetting"] = hotdataserve.GetSiteSettingsConfigCache()
	if err := ht4gooseforum.ExecuteTemplate(c.Writer, name, templateData); err != nil {
		slog.Error("render template err", "err", err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}

// GetTheme 从cookie中读取主题设置
func GetTheme(c *gin.Context) string {
	theme := "light" // 默认主题
	if themeCookie, err := c.Cookie("theme"); err == nil && themeCookie != "" {
		theme = themeCookie
	}
	return theme
}
