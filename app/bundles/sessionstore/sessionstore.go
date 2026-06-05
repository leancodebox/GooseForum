package sessionstore

import (
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
	})
	return store
}

func sessionSigningKey() string {
	if signingKey := preferences.GetString("app.signingKey"); signingKey != "" {
		return signingKey
	}
	return algorithm.SafeGenerateSigningKey(32)
}
