package component

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestLangCachesResultOnContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/?lang=en", nil)

	if got := RequestLang(c); got != "en" {
		t.Fatalf("RequestLang() = %q, want en", got)
	}

	c.Request = httptest.NewRequest("GET", "/?lang=ja", nil)
	if got := RequestLang(c); got != "en" {
		t.Fatalf("cached RequestLang() = %q, want en", got)
	}
}
