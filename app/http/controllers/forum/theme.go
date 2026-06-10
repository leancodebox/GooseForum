package forum

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/localcache"
	"github.com/leancodebox/GooseForum/app/models/defaultconfig"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
)

const (
	siteThemeRuntimeCacheTTL = 5 * time.Second
	siteThemeHistoryLimit    = 5
	siteThemeLightName       = "gf-light"
	siteThemeDarkName        = "gf-dark"
)

type runtimeSiteTheme struct {
	Config pageConfig.SiteThemeConfig
	Colors runtimeSiteThemeColors
	CSS    string
	ETag   string
	Href   string
}

type runtimeSiteThemeColors struct {
	Light string
	Dark  string
}

func (colors runtimeSiteThemeColors) Payload() map[string]string {
	if colors.Light == "" && colors.Dark == "" {
		return nil
	}
	payload := make(map[string]string, 2)
	if colors.Light != "" {
		payload[siteThemeLightName] = colors.Light
	}
	if colors.Dark != "" {
		payload[siteThemeDarkName] = colors.Dark
	}
	return payload
}

var siteThemeRuntimeCache = &localcache.Cache[runtimeSiteTheme]{MaxEntries: 1}

type ThemePreviewProps struct {
	Theme    pageConfig.SiteThemeConfig `json:"theme"`
	Defaults pageConfig.SiteThemeConfig `json:"defaults"`
}

func ThemePreview(c *gin.Context) {
	defaultTheme := defaultconfig.GetDefaultSiteThemeConfig()
	payload := PagePayload{
		Component: "theme.preview",
		Props: ThemePreviewProps{
			Theme:    normalizeSiteThemeConfig(pageConfig.GetConfigByPageType(pageConfig.SiteTheme, defaultTheme)),
			Defaults: normalizeSiteThemeConfig(defaultTheme),
		},
		Meta:    buildSimpleMeta(c, "主题预览"),
		Layout:  buildLayout(c, "theme-preview"),
		URL:     buildPageURL(c),
		Version: payloadVersion,
	}
	c.Header("Vary", "X-Goose-Page, Accept")
	c.Status(http.StatusOK)
	renderPage(c, "theme.gohtml", payload)
}

func SiteThemeCSS(c *gin.Context) {
	runtimeTheme := getRuntimeSiteTheme()
	if !runtimeTheme.Config.Enabled {
		c.Status(http.StatusNotFound)
		return
	}
	if runtimeTheme.CSS == "" {
		c.Status(http.StatusNotFound)
		return
	}

	c.Header("Content-Type", "text/css; charset=utf-8")
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Header("ETag", runtimeTheme.ETag)
	if match := c.GetHeader("If-None-Match"); match != "" && match == runtimeTheme.ETag {
		c.Status(http.StatusNotModified)
		return
	}
	c.String(http.StatusOK, runtimeTheme.CSS)
}

func getRuntimeSiteTheme() runtimeSiteTheme {
	data, _ := siteThemeRuntimeCache.GetOrLoadE("", func() (runtimeSiteTheme, error) {
		return buildRuntimeSiteTheme(hotdataserve.GetSiteThemeConfigCache()), nil
	}, siteThemeRuntimeCacheTTL)
	return data
}

func ClearSiteThemeRuntimeCache() {
	siteThemeRuntimeCache.Clear()
}

func buildRuntimeSiteTheme(rawConfig pageConfig.SiteThemeConfig) runtimeSiteTheme {
	config := normalizeSiteThemeConfig(rawConfig)
	css := buildSiteThemeCSS(config)
	colors := siteThemeColors(config)
	etag := ""
	if css != "" {
		sum := sha256.Sum256([]byte(css))
		etag = `"` + hex.EncodeToString(sum[:8]) + `"`
	}
	href := ""
	if config.Enabled && css != "" {
		version := config.PublishedAt
		if version == "" {
			version = strconv.Itoa(config.Version)
		}
		href = "/site-theme.css?v=" + url.QueryEscape(version)
	}
	return runtimeSiteTheme{
		Config: config,
		Colors: colors,
		CSS:    css,
		ETag:   etag,
		Href:   href,
	}
}

func normalizeSiteThemeConfig(config pageConfig.SiteThemeConfig) pageConfig.SiteThemeConfig {
	config = cloneSiteThemeConfig(config)
	defaultConfig := defaultconfig.GetDefaultSiteThemeConfig()
	fallbackTheme := pageConfig.FirstSiteThemeDefinition(defaultConfig.Themes)
	if config.Version <= 0 {
		config.Version = defaultConfig.Version
	}
	config.Themes = pageConfig.NormalizeSiteThemeDefinitions(config.Themes, defaultConfig.Themes, fallbackTheme)
	if config.Draft == nil {
		config.Draft = &pageConfig.SiteThemeSnapshot{
			Enabled: config.Enabled,
			Themes:  cloneSiteThemeDefinitions(config.Themes),
			Label:   "published",
		}
	}
	config.Draft.Themes = pageConfig.NormalizeSiteThemeDefinitions(config.Draft.Themes, defaultConfig.Themes, fallbackTheme)
	config.History = pageConfig.NormalizeSiteThemeSnapshots(config.History, defaultConfig.Themes, fallbackTheme, siteThemeHistoryLimit)
	return config
}

func cloneSiteThemeDefinitions(themes []pageConfig.SiteThemeDefinition) []pageConfig.SiteThemeDefinition {
	cloned := make([]pageConfig.SiteThemeDefinition, len(themes))
	copy(cloned, themes)
	return cloned
}

func cloneSiteThemeSnapshot(snapshot *pageConfig.SiteThemeSnapshot) *pageConfig.SiteThemeSnapshot {
	if snapshot == nil {
		return nil
	}
	cloned := *snapshot
	cloned.Themes = cloneSiteThemeDefinitions(snapshot.Themes)
	return &cloned
}

func cloneSiteThemeSnapshots(snapshots []pageConfig.SiteThemeSnapshot) []pageConfig.SiteThemeSnapshot {
	if snapshots == nil {
		return nil
	}
	cloned := make([]pageConfig.SiteThemeSnapshot, len(snapshots))
	for index, snapshot := range snapshots {
		cloned[index] = snapshot
		cloned[index].Themes = cloneSiteThemeDefinitions(snapshot.Themes)
	}
	return cloned
}

func cloneSiteThemeConfig(config pageConfig.SiteThemeConfig) pageConfig.SiteThemeConfig {
	config.Themes = cloneSiteThemeDefinitions(config.Themes)
	config.Draft = cloneSiteThemeSnapshot(config.Draft)
	config.History = cloneSiteThemeSnapshots(config.History)
	return config
}

func buildSiteThemeCSS(config pageConfig.SiteThemeConfig) string {
	if !config.Enabled {
		return ""
	}

	var sb strings.Builder
	for _, theme := range config.Themes {
		name := sanitizeThemeName(theme.Name)
		if name == "" {
			continue
		}
		sb.WriteString(`[data-theme="`)
		sb.WriteString(name)
		sb.WriteString(`"]{`)
		if theme.ColorScheme == "dark" || theme.ColorScheme == "light" {
			sb.WriteString("color-scheme:")
			sb.WriteString(theme.ColorScheme)
			sb.WriteByte(';')
		}
		theme.Tokens.AppendCSSVariables(&sb)
		sb.WriteString("}\n")
	}
	return sb.String()
}

func siteThemeColors(config pageConfig.SiteThemeConfig) runtimeSiteThemeColors {
	var colors runtimeSiteThemeColors
	if !config.Enabled {
		return colors
	}
	for _, theme := range config.Themes {
		name := sanitizeThemeName(theme.Name)
		if name == "" {
			continue
		}
		color := theme.Tokens.BaseColor()
		if color != "" {
			switch name {
			case siteThemeLightName:
				colors.Light = color
			case siteThemeDarkName:
				colors.Dark = color
			}
		}
	}
	return colors
}

func sanitizeThemeName(value string) string {
	value = strings.TrimSpace(value)
	if value != siteThemeLightName && value != siteThemeDarkName {
		return ""
	}
	return value
}
