package tokenservice

import "github.com/leancodebox/GooseForum/app/bundles/preferences"

func secretKey() []byte {
	return []byte(preferences.SecretKey())
}
