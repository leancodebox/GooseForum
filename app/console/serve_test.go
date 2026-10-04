package console

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
)

func TestForwardingHeadersRequireAnExplicitTrustedProxy(t *testing.T) {
	previous := preferences.GetStringSlice("server.trustedProxies")
	t.Cleanup(func() { preferences.Set("server.trustedProxies", previous) })
	for _, tc := range []struct {
		name    string
		proxies []string
		want    string
	}{
		{"untrusted", nil, "192.0.2.1"},
		{"trusted", []string{"192.0.2.0/24"}, "198.51.100.8"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			preferences.Set("server.trustedProxies", tc.proxies)
			engine, err := newGinEngine()
			if err != nil {
				t.Fatal(err)
			}
			engine.GET("/ip", func(c *gin.Context) { c.String(http.StatusOK, c.ClientIP()) })
			req := httptest.NewRequest(http.MethodGet, "/ip", nil)
			req.RemoteAddr = "192.0.2.1:2345"
			req.Header.Set("X-Forwarded-For", "198.51.100.8")
			r := httptest.NewRecorder()
			engine.ServeHTTP(r, req)
			if r.Body.String() != tc.want {
				t.Fatalf("client IP: %s", r.Body.String())
			}
		})
	}
	preferences.Set("server.trustedProxies", []string{"invalid proxy"})
	if _, err := newGinEngine(); err == nil {
		t.Fatal("invalid proxy configuration accepted")
	}
}
