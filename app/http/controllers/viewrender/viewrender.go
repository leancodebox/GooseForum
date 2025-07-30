package viewrender

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/http/controllers/datacache"
	"github.com/leancodebox/GooseForum/app/http/controllers/transform"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
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
			return webSettingsCache.GetOrLoad("websetcache", func() (pageConfig.WebSettingsConfig, error) {
				return pageConfig.GetConfigByPageType(pageConfig.WebSettings, pageConfig.WebSettingsConfig{}), nil
			},
				time.Second*5,
			)
		},
	}
}

func Render(c *gin.Context, name string, templateData map[string]any) {
	if templateData == nil {
		templateData = make(map[string]any, 4)
	}
	templateData["User"] = GetLoginUser(c)
	templateData["IsProduction"] = setting.IsProduction()
	templateData["Theme"] = GetTheme(c)
	templateData["Footer"] = hotdataserve.GetFooterConfigCache()
	templateData["SiteSetting"] = hotdataserve.GetSiteSettingsConfigCache()
	if _, ok := templateData["PageMeta"]; !ok {
		templateData["PageMeta"] = NewPageMetaBuilder().Build()
	}
	if err := ht4gooseforum.ExecuteTemplate(c.Writer, name, templateData); err != nil {
		slog.Error("render template err", "err", err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}

func GetLoginUser(c *gin.Context) *vo.UserInfoShow {
	userId := c.GetUint64("userId")
	return GetUserShowByUserId(userId)
}

func GetUserShowByUserId(userId uint64) *vo.UserInfoShow {
	if userId == 0 {
		return &vo.UserInfoShow{}
	}
	return hotdataserve.GetOrLoad(fmt.Sprintf("user:%v", userId), func() (*vo.UserInfoShow, error) {
		user, _ := users.Get(userId)
		if user.Id == 0 {
			return &vo.UserInfoShow{}, errors.New("no found")
		}
		return transform.User2userShow(user), nil
	})
}

type TmplData struct {
	IsProduction bool
	Theme        string
	Footer       pageConfig.FooterConfig
	SiteSetting  pageConfig.SiteSettingsConfig
	Data         map[string]any
}

func SafeRender(c *gin.Context, name string, data map[string]any) {
	templateData := TmplData{
		IsProduction: setting.IsProduction(),
		Theme:        GetTheme(c),
		Footer:       hotdataserve.GetFooterConfigCache(),
		SiteSetting:  hotdataserve.GetSiteSettingsConfigCache(),
		Data:         data,
	}
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
