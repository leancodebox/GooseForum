package tokenservice

import "github.com/leancodebox/GooseForum/app/bundles/preferences"

func signingKey() []byte {
	return []byte(preferences.GetString("app.signingKey"))
}
