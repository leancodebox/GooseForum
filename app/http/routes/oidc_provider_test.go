package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	core "github.com/leancodebox/GooseForum/app/bundles/oidcprovider"
	oidchttp "github.com/leancodebox/GooseForum/app/http/controllers/oidcprovider"
)

func TestOIDCProviderRoutesAreRegistered(t *testing.T) {
	handler, err := oidchttp.New(func() (*core.Provider, error) { return nil, context.Canceled }, nil)
	if err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	RegisterOIDCProvider(engine, handler)
	want := map[string]bool{
		"GET /oauth2/.well-known/openid-configuration": false,
		"GET /oauth2/authorize":                        false,
		"GET /oauth2/authorize/resume":                 false,
		"GET /oauth2/consent/details":                  false,
		"POST /oauth2/consent":                         false,
		"POST /oauth2/token":                           false,
		"GET /oauth2/userinfo":                         false,
		"POST /oauth2/userinfo":                        false,
		"GET /oauth2/jwks.json":                        false,
		"POST /oauth2/revoke":                          false,
	}
	for _, route := range engine.Routes() {
		key := route.Method + " " + route.Path
		if _, exists := want[key]; exists {
			want[key] = true
		}
	}
	for route, found := range want {
		if !found {
			t.Errorf("route not registered: %s", route)
		}
	}

	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/oauth2/jwks.json", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("registered handler status=%d", response.Code)
	}
}
