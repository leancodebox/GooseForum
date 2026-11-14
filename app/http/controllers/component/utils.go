package component

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
)

func GetBaseUri(c *gin.Context) string {
	scheme := "https"
	if strings.HasPrefix(c.Request.Host, "localhost") {
		scheme = "http"
	}
	host := fmt.Sprintf("%s://%s", scheme, c.Request.Host)
	return preferences.Get("server.url", host)
}

// GetHost is kept as a compatibility alias for callers that need the configured base URL.
func GetHost(c *gin.Context) string {
	return GetBaseUri(c)
}
