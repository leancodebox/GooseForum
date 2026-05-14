package viewrender

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	appI18n "github.com/leancodebox/GooseForum/app/service/i18n"
	"github.com/leancodebox/GooseForum/resource"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var (
	CurrentRegistry *TemplateRegistry
)

func Reload() {
	appI18n.Init(resource.GetTemplateFS())

	var err error
	CurrentRegistry, err = NewRegistry(resource.GetTemplateFS())
	if err != nil {
		slog.Error("Failed to reload templates", "err", err)
	}
}

func init() {
	Reload()
}

type TmplData[T any] struct {
	IsProduction  bool
	Theme         string
	Footer        pageConfig.FooterInfo
	SiteSetting   pageConfig.SiteSettingsConfig
	Data          T
	Url           URLHelper
	User          *vo.UserInfoShow
	CurrentUserId uint64
	PageMeta      *PageMeta
	Lang          string
	T             func(string, ...any) string
}

func SafeRender[T any](c *gin.Context, name string, data T, pageMeta ...*PageMeta) {
	user := component.GetLoginUser(c)
	var meta *PageMeta
	if len(pageMeta) > 0 && pageMeta[0] != nil {
		meta = pageMeta[0]
	} else {
		meta = NewPageMetaBuilder().Build()
	}

	lang := GetLang(c)

	localizer := appI18n.GetLocalizer(lang)
	tFunc := func(key string, args ...any) string {
		if localizer == nil {
			return key
		}

		config := &i18n.LocalizeConfig{
			MessageID: key,
		}
		if len(args) > 0 {
			config.TemplateData = args[0]
		}

		msg, err := localizer.Localize(config)
		if err != nil {
			return key
		}
		return msg
	}

	templateData := TmplData[T]{
		IsProduction:  setting.IsProduction(),
		Theme:         GetTheme(c),
		Footer:        hotdataserve.GetSiteSettingsConfigCache().FooterInfo,
		SiteSetting:   hotdataserve.GetSiteSettingsConfigCache(),
		Data:          data,
		Url:           URLHelper{},
		User:          user,
		Lang:          lang,
		CurrentUserId: component.LoginUserId(c),
		PageMeta:      meta,
		T:             tFunc,
	}

	if CurrentRegistry == nil {
		slog.Error("CurrentRegistry is nil")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if err := CurrentRegistry.Render(c.Writer, name, templateData); err != nil {
		slog.Error("render template err", "err", err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}

func GetLang(c *gin.Context) string {
	lang := c.Query("lang")
	if lang == "" {
		if cookie, err := c.Cookie("lang"); err == nil && cookie != "" {
			lang = cookie
		} else {
			lang = c.GetHeader("Accept-Language")
		}
	}
	if lang == "" {
		lang = "zh"
	}
	return lang
}

// GetTheme reads the selected theme from cookies.
func GetTheme(c *gin.Context) string {
	theme := "light"
	if themeCookie, err := c.Cookie("theme"); err == nil && themeCookie != "" {
		theme = themeCookie
	}
	return theme
}
