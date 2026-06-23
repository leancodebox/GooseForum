package sessionstore

import (
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/sessions"
	"github.com/leancodebox/GooseForum/app/bundles/algorithm"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
)

var store *sessions.CookieStore
var once sync.Once

func GetSession() *sessions.CookieStore {
	once.Do(func() {
		store = sessions.NewCookieStore([]byte(sessionSigningKey()))
		configureSessionStore(store)
	})
	return store
}

func sessionSigningKey() string {
	if signingKey := preferences.GetString("app.signingKey"); signingKey != "" {
		return signingKey
	}
	return algorithm.SafeGenerateSigningKey(32)
}

func configureSessionStore(store *sessions.CookieStore) {
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.SameSite = http.SameSiteLaxMode
	store.Options.Secure = sessionCookieSecure(preferences.GetString("server.url", ""), preferences.GetString("app.env", "production"))
}

func sessionCookieSecure(serverURL string, appEnv string) bool {
	normalizedURL := strings.ToLower(strings.TrimSpace(serverURL))
	if strings.HasPrefix(normalizedURL, "https://") {
		return true
	}
	if strings.HasPrefix(normalizedURL, "http://") {
		return false
	}
	return !strings.EqualFold(strings.TrimSpace(appEnv), "local")
}
