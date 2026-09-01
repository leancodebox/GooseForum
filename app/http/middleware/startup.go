package middleware

import (
	_ "embed"
	"net/http"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

//go:embed startup.html
var startupHTML []byte

// StartupGate keeps requests on a temporary page until startup work completes.
type StartupGate struct {
	ready atomic.Int32
}

func NewStartupGate() *StartupGate {
	return &StartupGate{}
}

func (g *StartupGate) Complete() {
	g.ready.Store(1)
}

func (g *StartupGate) Handler(c *gin.Context) {
	if g.ready.Load() == 1 {
		c.Next()
		return
	}
	c.Header("Retry-After", "5")
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
	c.Data(http.StatusServiceUnavailable, "text/html; charset=utf-8", startupHTML)
	c.Abort()
}
