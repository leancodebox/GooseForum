package middleware

import (
	"mime"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
)

// Token management accepts same-origin JSON commands, never cross-site forms.
func AgentTokenManagement(c *gin.Context) {
	c.Header("Cache-Control", "private, no-store")
	mediaType, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
	origin := c.GetHeader("Origin")
	allowed := origin == ""
	if parsed, parseErr := url.Parse(origin); origin != "" && parseErr == nil {
		allowed = (parsed.Scheme == "https" || parsed.Scheme == "http") && parsed.Host == c.Request.Host && parsed.Path == "" && parsed.RawQuery == "" && parsed.Fragment == ""
		if setting.IsLocal() && origin == preferences.GetString("resource.reactDevServer") {
			allowed = true
		}
	}
	if err != nil || mediaType != "application/json" || !allowed || c.GetHeader("Sec-Fetch-Site") == "cross-site" {
		c.AbortWithStatusJSON(http.StatusForbidden, component.FailDataCode(component.MessagePermissionDenied, nil))
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<16)
	c.Next()
}
