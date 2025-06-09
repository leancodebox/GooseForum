package databaseinit

import (
	"database/sql"
	_ "embed"
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"log"
	"strings"
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
		fmt.Println("不支持")
		return
	}

	db, err := sql.Open("mysql", dbUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sqlItem := strings.Split(dataAllSql, ";")
	for _, s := range sqlItem {
		_, err := db.Exec(s)
		if err != nil {
			fmt.Println(s)
			fmt.Println(err)
		}
	}
}
