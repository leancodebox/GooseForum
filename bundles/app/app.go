// Package app 应用信息
package app

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
