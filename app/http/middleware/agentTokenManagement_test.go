package middleware

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAgentTokenManagementRejectsCrossOriginAndFormCommands(t *testing.T) {
	for _, tc := range []struct {
		origin, contentType, site string
		want                      int
	}{
		{"https://forum.example", "application/json", "same-origin", 200},
		{"https://attacker.example", "application/json", "cross-site", 403},
		{"", "application/x-www-form-urlencoded", "", 403},
		{"null", "application/json", "", 403},
	} {
		router := gin.New()
		router.POST("/tokens", AgentTokenManagement, func(c *gin.Context) { c.Status(200) })
		req := httptest.NewRequest("POST", "https://forum.example/tokens", strings.NewReader("{}"))
		req.Header.Set("Origin", tc.origin)
		req.Header.Set("Content-Type", tc.contentType)
		req.Header.Set("Sec-Fetch-Site", tc.site)
		out := httptest.NewRecorder()
		router.ServeHTTP(out, req)
		if out.Code != tc.want {
			t.Errorf("origin %s content %s: %d, want %d", tc.origin, tc.contentType, out.Code, tc.want)
		}
	}
}
