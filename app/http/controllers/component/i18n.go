package component

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/i18n"
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
	lang := c.Query("lang")
	if lang == "" {
		if cookie, err := c.Cookie("lang"); err == nil && cookie != "" {
			lang = cookie
		} else {
			lang = c.GetHeader("Accept-Language")
		}
	}
	lang = i18n.Normalize(lang)
	c.Set(requestLangContextKey, lang)
	return lang
}
