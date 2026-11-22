package component

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/i18n"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"golang.org/x/text/language"
)

const requestLangContextKey = "goose:request-lang"

// RequestLang resolves the request locale (?lang -> "lang" cookie ->
// Accept-Language) and normalizes it to a supported locale. It is the single
// source of truth shared by the server-rendered templates and the account
// activation page, mirroring the frontend detectLocale().
func RequestLang(c *gin.Context) string {
	if cached, ok := c.Get(requestLangContextKey); ok {
		if lang, ok := cached.(string); ok && lang != "" {
			return lang
		}
	}
	lang := i18n.Supported(c.Query("lang"))
	if lang == "" {
		cookie, _ := c.Cookie("lang")
		lang = i18n.Supported(cookie)
	}
	if lang == "" {
		tags, _, _ := language.ParseAcceptLanguage(c.GetHeader("Accept-Language"))
		for _, tag := range tags {
			if lang = i18n.Supported(tag.String()); lang != "" {
				break
			}
		}
	}
	if lang == "" {
		lang = i18n.Normalize(hotdataserve.GetSiteSettingsConfigCache().DefaultLocale)
	}
	c.Set(requestLangContextKey, lang)
	return lang
}
