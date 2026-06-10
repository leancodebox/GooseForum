package forum

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/models/defaultconfig"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
)

type siteThemeCSSCacheEntry struct {
	key  string
	css  string
	etag string
}

var siteThemeCSSCache struct {
	sync.RWMutex
	entry siteThemeCSSCacheEntry
}

var allowedThemeTokens = map[string]bool{
	"color-base-100":          true,
	"color-base-200":          true,
	"color-base-300":          true,
	"color-base-content":      true,
	"color-icon-muted":        true,
	"color-line":              true,
	"color-primary":           true,
	"color-primary-content":   true,
	"color-secondary":         true,
	"color-secondary-content": true,
	"color-accent":            true,
	"color-accent-content":    true,
	"color-neutral":           true,
	"color-neutral-content":   true,
	"color-info":              true,
	"color-info-content":      true,
	"color-success":           true,
	"color-success-content":   true,
	"color-warning":           true,
	"color-warning-content":   true,
	"color-error":             true,
	"color-error-content":     true,
	"radius-selector":         true,
	"radius-field":            true,
	"radius-box":              true,
	"size-selector":           true,
	"size-field":              true,
	"border":                  true,
	"depth":                   true,
	"noise":                   true,
}

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
	config := normalizeSiteThemeConfig(hotdataserve.GetSiteThemeConfigCache())
	if !config.Enabled {
		c.Status(http.StatusNotFound)
		return
	}

	cached := getSiteThemeCSS(config)
	if cached.css == "" {
		c.Status(http.StatusNotFound)
		return
	}

	c.Header("Content-Type", "text/css; charset=utf-8")
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Header("ETag", cached.etag)
	if match := c.GetHeader("If-None-Match"); match != "" && match == cached.etag {
		c.Status(http.StatusNotModified)
		return
	}
	c.String(http.StatusOK, cached.css)
}

func normalizeSiteThemeConfig(config pageConfig.SiteThemeConfig) pageConfig.SiteThemeConfig {
	defaultConfig := defaultconfig.GetDefaultSiteThemeConfig()
	if config.Version <= 0 {
		config.Version = defaultConfig.Version
	}
	if len(config.Themes) == 0 {
		config.Themes = defaultConfig.Themes
	}

	defaultThemes := map[string]pageConfig.SiteThemeDefinition{}
	for _, theme := range defaultConfig.Themes {
		defaultThemes[theme.Name] = theme
	}

	for index := range config.Themes {
		theme := &config.Themes[index]
		if theme.Name == "" {
			theme.Name = defaultConfig.Themes[0].Name
		}
		defaultTheme := defaultThemes[theme.Name]
		if theme.Label == "" {
			theme.Label = defaultTheme.Label
		}
		if theme.ColorScheme != "dark" && theme.ColorScheme != "light" {
			theme.ColorScheme = defaultTheme.ColorScheme
		}
		if theme.Tokens == nil {
			theme.Tokens = map[string]string{}
		}
		for key, value := range defaultTheme.Tokens {
			if strings.TrimSpace(theme.Tokens[key]) == "" {
				theme.Tokens[key] = value
			}
		}
		for key := range theme.Tokens {
			if !allowedThemeTokens[key] {
				delete(theme.Tokens, key)
				continue
			}
			theme.Tokens[key] = normalizeLegacySiteThemeToken(key, theme.Tokens[key])
		}
	}
	if config.Draft == nil {
		config.Draft = &pageConfig.SiteThemeSnapshot{
			Enabled: config.Enabled,
			Themes:  cloneSiteThemeDefinitions(config.Themes),
			Label:   "published",
		}
	}
	config.Draft.Themes = normalizeSiteThemeDefinitions(config.Draft.Themes, defaultConfig, defaultThemes)
	if len(config.History) > 5 {
		config.History = config.History[len(config.History)-5:]
	}
	for index := range config.History {
		config.History[index].Themes = normalizeSiteThemeDefinitions(config.History[index].Themes, defaultConfig, defaultThemes)
	}

	return config
}

func normalizeSiteThemeDefinitions(themes []pageConfig.SiteThemeDefinition, defaultConfig pageConfig.SiteThemeConfig, defaultThemes map[string]pageConfig.SiteThemeDefinition) []pageConfig.SiteThemeDefinition {
	if len(themes) == 0 {
		themes = cloneSiteThemeDefinitions(defaultConfig.Themes)
	}
	for index := range themes {
		theme := &themes[index]
		if theme.Name == "" {
			theme.Name = defaultConfig.Themes[0].Name
		}
		defaultTheme := defaultThemes[theme.Name]
		if theme.Label == "" {
			theme.Label = defaultTheme.Label
		}
		if theme.ColorScheme != "dark" && theme.ColorScheme != "light" {
			theme.ColorScheme = defaultTheme.ColorScheme
		}
		if theme.Tokens == nil {
			theme.Tokens = map[string]string{}
		}
		for key, value := range defaultTheme.Tokens {
			if strings.TrimSpace(theme.Tokens[key]) == "" {
				theme.Tokens[key] = value
			}
		}
		for key := range theme.Tokens {
			if !allowedThemeTokens[key] {
				delete(theme.Tokens, key)
				continue
			}
			theme.Tokens[key] = normalizeLegacySiteThemeToken(key, theme.Tokens[key])
		}
	}
	return themes
}

func normalizeLegacySiteThemeToken(key string, value string) string {
	if key == "radius-field" {
		switch strings.TrimSpace(value) {
		case "0.375rem", "6px":
			return "0.5rem"
		}
	}
	return value
}

func cloneSiteThemeDefinitions(themes []pageConfig.SiteThemeDefinition) []pageConfig.SiteThemeDefinition {
	cloned := make([]pageConfig.SiteThemeDefinition, len(themes))
	for index, theme := range themes {
		cloned[index] = theme
		cloned[index].Tokens = map[string]string{}
		for key, value := range theme.Tokens {
			cloned[index].Tokens[key] = value
		}
	}
	return cloned
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
		keys := make([]string, 0, len(theme.Tokens))
		for key := range theme.Tokens {
			if allowedThemeTokens[key] {
				keys = append(keys, key)
			}
		}
		sort.Strings(keys)
		for _, key := range keys {
			value := sanitizeThemeValue(theme.Tokens[key])
			if value == "" {
				continue
			}
			sb.WriteString("--gf-")
			sb.WriteString(key)
			sb.WriteByte(':')
			sb.WriteString(value)
			sb.WriteByte(';')
		}
		sb.WriteString("}\n")
	}
	return sb.String()
}

func siteThemeHref(config pageConfig.SiteThemeConfig) string {
	if !config.Enabled {
		return ""
	}
	cached := getSiteThemeCSS(config)
	if cached.css == "" {
		return ""
	}
	version := config.PublishedAt
	if version == "" {
		version = strconv.Itoa(config.Version)
	}
	return "/site-theme.css?v=" + url.QueryEscape(version)
}

func siteThemeColors(config pageConfig.SiteThemeConfig) map[string]string {
	colors := map[string]string{}
	if !config.Enabled {
		return colors
	}
	for _, theme := range config.Themes {
		name := sanitizeThemeName(theme.Name)
		if name == "" {
			continue
		}
		color := sanitizeThemeValue(theme.Tokens["color-base-100"])
		if color != "" {
			colors[name] = color
		}
	}
	return colors
}

func getSiteThemeCSS(config pageConfig.SiteThemeConfig) siteThemeCSSCacheEntry {
	key := siteThemeCSSCacheKey(config)

	siteThemeCSSCache.RLock()
	if siteThemeCSSCache.entry.key == key {
		entry := siteThemeCSSCache.entry
		siteThemeCSSCache.RUnlock()
		return entry
	}
	siteThemeCSSCache.RUnlock()

	css := buildSiteThemeCSS(config)
	sum := sha256.Sum256([]byte(css))
	entry := siteThemeCSSCacheEntry{
		key:  key,
		css:  css,
		etag: `"` + hex.EncodeToString(sum[:8]) + `"`,
	}

	siteThemeCSSCache.Lock()
	siteThemeCSSCache.entry = entry
	siteThemeCSSCache.Unlock()
	return entry
}

func siteThemeCSSCacheKey(config pageConfig.SiteThemeConfig) string {
	payload := struct {
		Enabled     bool                             `json:"enabled"`
		Themes      []pageConfig.SiteThemeDefinition `json:"themes"`
		PublishedAt string                           `json:"publishedAt,omitempty"`
	}{
		Enabled:     config.Enabled,
		Themes:      config.Themes,
		PublishedAt: config.PublishedAt,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return config.PublishedAt
	}
	return string(data)
}

func sanitizeThemeName(value string) string {
	value = strings.TrimSpace(value)
	if value != "gf-light" && value != "gf-dark" {
		return ""
	}
	return value
}

func sanitizeThemeValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || strings.ContainsAny(value, "{};<>") {
		return ""
	}
	return value
}
