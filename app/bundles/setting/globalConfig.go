// Package app 应用信息
package setting

import (
	"github.com/leancodebox/goose/preferences"
)

const on = "on"
const off = "off"

func UseMigration() bool {
	openMigration := preferences.GetString("db.migration")
	return openMigration == on
}

func IsProduction() bool {
	return preferences.Get("app.env", "local") == "production"
}

// URL 传参 path 拼接站点的 URL
func URL(path string) string {
	return preferences.Get("app.url") + path
}
