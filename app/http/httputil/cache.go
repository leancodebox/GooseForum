// Package httputil contains small HTTP response helpers.
package httputil

import (
	"github.com/gin-gonic/gin"
)

// SetLongPublic marks the response as publicly cacheable for static assets.
func SetLongPublic(c *gin.Context) {
	c.Header("Cache-Control", "public, max-age=18144000")
}
