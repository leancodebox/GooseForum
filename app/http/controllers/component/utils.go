package component

import (
	"cmp"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
)

var (
	minPageSize int = 10
	maxPageSize int = 30
)

func BoundPageSize(pageSize int) int {
	return BoundPageSizeWithRange(pageSize, minPageSize, maxPageSize)
}

func BoundPageSizeWithRange[T cmp.Ordered](pageSize T, minN, maxN T) T {
	return min(max(pageSize, minN), maxN)
}

func BuildCanonicalHref(c *gin.Context) string {
	return GetBaseUri(c) + c.Request.URL.String()
}

func GetBaseUri(c *gin.Context) string {
	scheme := "https"
	if strings.HasPrefix(c.Request.Host, "localhost") {
		scheme = "http"
	}
	host := fmt.Sprintf("%s://%s", scheme, c.Request.Host)
	return preferences.Get("server.url", host)
}

func GetHost(c *gin.Context) string {
	scheme := "https"
	if strings.HasPrefix(c.Request.Host, "localhost") {
		scheme = "http"
	}
	host := fmt.Sprintf("%s://%s", scheme, c.Request.Host)
	return preferences.Get("server.url", host)
}
