package controllers

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
)

func TestOutboundPolicy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, target string
		enabled      bool
		whitelist    []string
		status       int
	}{
		{"disabled", "https://other.test/a?q=x#part", false, nil, 302},
		{"warning", "https://other.test/a?q=x#part", true, nil, 200},
		{"allowed", "https://other.test/a", true, []string{"other.test"}, 302},
		{"suffix attack", "https://other.test.evil.test", true, []string{"other.test"}, 200},
		{"same site", "https://forum.test/p/post/1", true, nil, 302},
		{"script", "javascript:alert(1)", false, nil, 400},
		{"credentials", "https://forum.test@evil.test", true, nil, 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			engine := gin.New()
			engine.GET("/outbound", func(c *gin.Context) {
				serveOutbound(c, tc.enabled, tc.whitelist, "https://forum.test", "GooseForum", pageConfig.SiteChromeConfig{})
			})
			response := httptest.NewRecorder()
			engine.ServeHTTP(response, httptest.NewRequest("GET", "/outbound?url="+url.QueryEscape(tc.target), nil))
			if response.Code != tc.status {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
			if tc.status == 302 && response.Header().Get("Location") != tc.target {
				t.Fatal("destination changed")
			}
			if tc.status == 200 && (!strings.Contains(response.Body.String(), "继续访问") || response.Header().Get("Location") != "") {
				t.Fatal("missing warning")
			}
			if response.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("policy must not be cached")
			}
		})
	}
}

func TestOutboundEscapesDestination(t *testing.T) {
	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Request = httptest.NewRequest("GET", "/outbound?url="+url.QueryEscape("https://other.test/?q=<script>alert(1)</script>"), nil)
	serveOutbound(c, true, nil, "https://forum.test", "<script>bad</script>", pageConfig.SiteChromeConfig{})
	if strings.Contains(response.Body.String(), "<script>") {
		t.Fatal("unescaped content")
	}
}

func TestOutboundUsesConfiguredBrand(t *testing.T) {
	for _, tc := range []struct {
		chrome pageConfig.SiteChromeConfig
		want   string
	}{
		{pageConfig.SiteChromeConfig{BrandType: "text", BrandText: "社区名称"}, "社区名称"},
		{pageConfig.SiteChromeConfig{BrandType: "image", BrandImage: "/file/logo.svg"}, `src="/file/logo.svg"`},
	} {
		response := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(response)
		c.Request = httptest.NewRequest("GET", "/outbound?url=https%3A%2F%2Fother.test", nil)
		serveOutbound(c, true, nil, "https://forum.test", "站点名称", tc.chrome)
		if !strings.Contains(response.Body.String(), tc.want) {
			t.Fatal(response.Body.String())
		}
	}
}
