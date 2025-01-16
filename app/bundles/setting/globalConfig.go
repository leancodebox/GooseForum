// Package app 应用信息
package setting

import (
	"github.com/leancodebox/GooseForum/app/bundles/goose/preferences"
)

const on = "on"
const off = "off"

func UseMigration() bool {
	openMigration := preferences.GetString("db.migration")
	return openMigration == on
}

func IsProduction() bool {
	return preferences.Get("app.env", "production") == "production"
}

func IsDebug() bool {
	return preferences.GetBool("app.debug", true)
}

// URL 传参 path 拼接站点的 URL
func URL(path string) string {
	return preferences.Get("server.url") + path
}
