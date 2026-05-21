package httputil

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func SetLongPublic(c *gin.Context) {
	c.Header("Cache-Control", "public, max-age=18144000")
}

func SetNoStore(c *gin.Context) {
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
}

func IsAdminIndexPath(path string) bool {
	cleanPath := strings.TrimRight(path, "/")
	return cleanPath == "/admin" || cleanPath == "/admin/index.html"
}
