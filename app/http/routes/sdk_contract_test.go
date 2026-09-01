package routes

import (
	"os"
	"regexp"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPublicThemeSDKRoutesAreRegistered(t *testing.T) {
	source, err := os.ReadFile("../../../resource/packages/client/src/api/routes.ts")
	if err != nil {
		t.Fatalf("read SDK route catalog: %v", err)
	}
	pattern := regexp.MustCompile(`\w+: \['(GET|POST)', '([^']+)'\]`)
	matches := pattern.FindAllStringSubmatch(string(source), -1)
	if len(matches) == 0 {
		t.Fatal("SDK route catalog contained no routes")
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	apiRoute(router)
	fileServer(router)
	registered := make(map[string]bool, len(router.Routes()))
	for _, route := range router.Routes() {
		registered[route.Method+" "+route.Path] = true
	}

	for _, match := range matches {
		method, path := match[1], match[2]
		if !registered[method+" "+path] {
			t.Errorf("SDK route is not registered by the Go server: %s %s", method, path)
		}
	}
}
