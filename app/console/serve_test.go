package console

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestVisitorIP(t *testing.T) {
	engine, err := newGinEngine()
	if err != nil {
		t.Fatal(err)
	}
	engine.GET("/ip", func(c *gin.Context) { c.String(http.StatusOK, c.ClientIP()) })

	for _, tc := range []struct {
		name, peer, realIP, forwardedFor, want string
	}{
		{"direct", "192.0.2.1:1234", "", "", "192.0.2.1"},
		{"direct with forwarding headers", "192.0.2.1:1234", "198.51.100.8", "203.0.113.9", "192.0.2.1"},
		{"container gateway", "172.17.0.1:1234", "198.51.100.8", "", "172.17.0.1"},
		{"local proxy", "127.0.0.1:1234", "198.51.100.8", "203.0.113.9", "198.51.100.8"},
		{"IPv6 local proxy", "[::1]:1234", "2001:db8::8", "", "2001:db8::8"},
		{"missing real IP", "127.0.0.1:1234", "", "203.0.113.9", "127.0.0.1"},
		{"invalid real IP", "127.0.0.1:1234", "unknown", "", "127.0.0.1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/ip", nil)
			req.RemoteAddr = tc.peer
			req.Header.Set("X-Real-IP", tc.realIP)
			req.Header.Set("X-Forwarded-For", tc.forwardedFor)
			r := httptest.NewRecorder()
			engine.ServeHTTP(r, req)
			if got := r.Body.String(); got != tc.want {
				t.Fatalf("client IP = %q, want %q", got, tc.want)
			}
		})
	}
}
