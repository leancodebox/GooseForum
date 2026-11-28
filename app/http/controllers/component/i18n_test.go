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

func TestRequestLangPriority(t *testing.T) {
	for _, tc := range []struct{ query, cookie, header, want string }{
		{"it", "en", "ja", "it"},
		{"nl", "en", "ja", "en"},
		{"nl", "nl", "nl-NL, en;q=0.3, ja-JP;q=0.8", "ja"},
		{"ko-KR", "en", "ja", "ko"},
		{"", "fr-CA", "en", "fr"},
		{"", "", "nl-NL, fr-FR;q=0.9", "fr"},
		{"", "", "zh;q=0, it;q=0.5", "it"},
	} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/?lang="+tc.query, nil)
		c.Request.Header.Set("Cookie", "lang="+tc.cookie)
		c.Request.Header.Set("Accept-Language", tc.header)
		if got := RequestLang(c); got != tc.want {
			t.Errorf("%+v: got %q", tc, got)
		}
	}
}
