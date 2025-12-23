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
	"github.com/leancodebox/GooseForum/resource"
)

var (
	CurrentRegistry *TemplateRegistry
)

func Reload() {
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
	Footer        pageConfig.FooterConfig
	SiteSetting   pageConfig.SiteSettingsConfig
	Data          T
	Url           resource.URLHelper
	User          *vo.UserInfoShow
	CurrentUserId uint64
	PageMeta      *PageMeta
}

func SafeRender[T any](c *gin.Context, name string, data T, pageMeta ...*PageMeta) {
	user := component.GetLoginUser(c)
	var meta *PageMeta
	if len(pageMeta) > 0 && pageMeta[0] != nil {
		meta = pageMeta[0]
	} else {
		meta = NewPageMetaBuilder().Build()
	}

	templateData := TmplData[T]{
		IsProduction:  setting.IsProduction(),
		Theme:         GetTheme(c),
		Footer:        hotdataserve.GetFooterConfigCache(),
		SiteSetting:   hotdataserve.GetSiteSettingsConfigCache(),
		Data:          data,
		Url:           resource.URLHelper{},
		User:          user,
		CurrentUserId: user.UserId,
		PageMeta:      meta,
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

// GetTheme 从cookie中读取主题设置
func GetTheme(c *gin.Context) string {
	theme := "light" // 默认主题
	if themeCookie, err := c.Cookie("theme"); err == nil && themeCookie != "" {
		theme = themeCookie
	}
	return theme
}
