package databaseinit

import (
	"database/sql"
	_ "embed"
	"log/slog"
	"strings"

	"github.com/leancodebox/GooseForum/app/bundles/preferences"
)

//go:embed dbinit/init.sql
var dataAllSql string

func DBInit() {
	var (
		dbConfig   = preferences.GetExclusivePreferences("db.default")
		connection = dbConfig.Get(`connection`, `sqlite`)
		dbUrl      = dbConfig.Get(`url`)
	)
	if connection == "sqlite" {
		slog.Info("database init skipped for sqlite")
		return
	}

	db, err := sql.Open("mysql", dbUrl)
	if err != nil {
		slog.Error("init err", "err", err)
		return
	}
	defer db.Close()

	sqlItem := strings.SplitSeq(dataAllSql, ";")
	for s := range sqlItem {
		_, err := db.Exec(s)
		if err != nil {
			slog.Error("database init statement failed", "sql", s, "error", err)
		}
	}
}
