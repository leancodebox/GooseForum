package forum

import (
	"html/template"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/i18n"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
)

var templateFuncMap = template.FuncMap{
	"SafeHTML": func(s string) template.HTML {
		return template.HTML(s)
	},
	"Nl2br": func(text string) template.HTML {
		escaped := template.HTMLEscapeString(text)
		result := strings.ReplaceAll(escaped, "\n", "<br>")
		return template.HTML(result)
	},
	"json": func(v any) template.JS {
		return template.JS(jsonopt.Encode(v))
	},
	// t localizes a server-rendered string, e.g. {{ t .Lang "search" }}.
	// Extra args are alternating name/value pairs for {name} placeholders.
	"t": i18n.T,
}

// requestLang resolves and normalizes the request locale. It delegates to
// component.RequestLang so the templates and the activation page share one
// source of truth. The normalized value is also safe as an <html lang> value.
func requestLang(c *gin.Context) string {
	return component.RequestLang(c)
}
