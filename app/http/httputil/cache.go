package httputil

import (
	"github.com/gin-gonic/gin"
)

func SetLongPublic(c *gin.Context) {
	c.Header("Cache-Control", "public, max-age=18144000")
}
