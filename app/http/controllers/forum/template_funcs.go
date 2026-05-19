package forum

import (
	"html/template"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
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
}

func requestLang(c *gin.Context) string {
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
