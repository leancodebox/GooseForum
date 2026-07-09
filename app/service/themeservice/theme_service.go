package themeservice

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/localcache"
	"github.com/leancodebox/GooseForum/app/models/defaultconfig"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
)

const (
	runtimeCacheTTL = 5 * time.Second
	LightName       = "gf-light"
	DarkName        = "gf-dark"
)

type RuntimeTheme struct {
	Enabled bool
	Colors  RuntimeThemeColors
	CSS     string
	ETag    string
	Href    string
}

type RuntimeThemeColors struct {
	Light string
	Dark  string
}

func (colors RuntimeThemeColors) Payload() map[string]string {
	if colors.Light == "" && colors.Dark == "" {
		return nil
	}
	payload := make(map[string]string, 2)
	if colors.Light != "" {
		payload[LightName] = colors.Light
	}
	if colors.Dark != "" {
		payload[DarkName] = colors.Dark
	}
	return payload
}

var runtimeCache = &localcache.Cache[RuntimeTheme]{MaxEntries: 1}

func Defaults() pageConfig.SiteThemeConfig {
	return NormalizeConfig(defaultconfig.GetDefaultSiteThemeConfig())
}

func LoadConfig() pageConfig.SiteThemeConfig {
	return NormalizeConfig(pageConfig.GetConfigByPageType(pageConfig.SiteTheme, defaultconfig.GetDefaultSiteThemeConfig()))
}

func Runtime() RuntimeTheme {
	return runtimeCache.GetOrLoad("", func() (RuntimeTheme, error) {
		return BuildRuntime(hotdataserve.GetSiteThemeConfigCache()), nil
	}, runtimeCacheTTL)
}

func ClearCaches() {
	hotdataserve.ClearSiteThemeConfigCache()
	runtimeCache.Clear()
}

func BuildRuntime(rawConfig pageConfig.SiteThemeConfig) RuntimeTheme {
	config := NormalizeConfig(rawConfig)
	css := BuildCSS(config)
	colors := themeColors(config)
	hash := ""
	etag := ""
	if css != "" {
		sum := sha256.Sum256([]byte(css))
		hash = hex.EncodeToString(sum[:8])
		etag = `"` + hash + `"`
	}
	href := ""
	if config.Enabled && css != "" {
		href = "/site-theme.css?v=" + url.QueryEscape(hash)
	}
	return RuntimeTheme{
		Enabled: config.Enabled,
		Colors:  colors,
		CSS:     css,
		ETag:    etag,
		Href:    href,
	}
}

func NormalizeConfig(config pageConfig.SiteThemeConfig) pageConfig.SiteThemeConfig {
	config = CloneConfig(config)
	defaultConfig := defaultconfig.GetDefaultSiteThemeConfig()
	fallbackTheme := pageConfig.FirstSiteThemeDefinition(defaultConfig.Themes)
	if config.Version <= 0 {
		config.Version = defaultConfig.Version
	}
	config.Themes = pageConfig.NormalizeSiteThemeDefinitions(config.Themes, defaultConfig.Themes, fallbackTheme)
	config.Prepublish = NormalizePrepublish(config.Prepublish, defaultConfig.Themes, fallbackTheme)
	return config
}

func CloneDefinitions(themes []pageConfig.SiteThemeDefinition) []pageConfig.SiteThemeDefinition {
	cloned := make([]pageConfig.SiteThemeDefinition, len(themes))
	copy(cloned, themes)
	return cloned
}

func NormalizePrepublish(prepublish *pageConfig.SiteThemePrepublish, defaults []pageConfig.SiteThemeDefinition, fallback pageConfig.SiteThemeDefinition) *pageConfig.SiteThemePrepublish {
	if prepublish == nil || len(prepublish.Themes) == 0 {
		return nil
	}
	prepublish = ClonePrepublish(prepublish)
	prepublish.Themes = pageConfig.NormalizeSiteThemeDefinitions(prepublish.Themes, defaults, fallback)
	return prepublish
}

func ClonePrepublish(prepublish *pageConfig.SiteThemePrepublish) *pageConfig.SiteThemePrepublish {
	if prepublish == nil {
		return nil
	}
	cloned := *prepublish
	cloned.Themes = CloneDefinitions(prepublish.Themes)
	return &cloned
}

func CloneConfig(config pageConfig.SiteThemeConfig) pageConfig.SiteThemeConfig {
	config.Themes = CloneDefinitions(config.Themes)
	config.Prepublish = ClonePrepublish(config.Prepublish)
	return config
}

func BuildCSS(config pageConfig.SiteThemeConfig) string {
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

func themeColors(config pageConfig.SiteThemeConfig) RuntimeThemeColors {
	var colors RuntimeThemeColors
	if !config.Enabled {
		return colors
	}
	for _, theme := range config.Themes {
		name := sanitizeThemeName(theme.Name)
		if name == "" {
			continue
		}
		color := theme.Tokens.BaseColor()
		if color == "" {
			continue
		}
		switch name {
		case LightName:
			colors.Light = color
		case DarkName:
			colors.Dark = color
		}
	}
	return colors
}

func sanitizeThemeName(value string) string {
	value = strings.TrimSpace(value)
	if value != LightName && value != DarkName {
		return ""
	}
	return value
}
