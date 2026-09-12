package forum

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLoginInitialMode(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{name: "default", url: "/login", want: "login"},
		{name: "register", url: "/login?mode=register", want: "register"},
		{name: "forgot", url: "/login?mode=forgot", want: "forgot"},
		{name: "legacy register", url: "/login?register=true", want: "register"},
		{name: "legacy model", url: "/login?model=register", want: "register"},
	}

	gin.SetMode(gin.TestMode)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _ := gin.CreateTestContext(httptest.NewRecorder())
			context.Request = httptest.NewRequest("GET", test.url, nil)
			if got := loginInitialMode(context); got != test.want {
				t.Fatalf("loginInitialMode() = %q, want %q", got, test.want)
			}
		})
	}
}
