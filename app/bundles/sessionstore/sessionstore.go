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
		secretKey := preferences.GetString("app.signingKey", algorithm.SafeGenerateSigningKey(32))
		store = sessions.NewCookieStore([]byte(secretKey))
	})
	return store
}
