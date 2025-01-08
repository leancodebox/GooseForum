package bbsinit

import (
	"database/sql"
	_ "embed"
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/goose/preferences"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategory"
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

func DataInit() {
	dataList := []articleCategory.Entity{
		{Category: "Code"},
		{Category: "Go"},
		{Category: "PHP"},
		{Category: "Python"},
		{Category: "Web"},
		{Category: "Server"},
		{Category: "App"},
		{Category: "Vue.js"},
		{Category: "Java"},
		{Category: "MySQL"},
		{Category: "Rust"},
		{Category: "Server"},
		{Category: "Database"},
		{Category: "Math"},
	}
	has := articleCategory.Get(1)
	if has.Id == 0 {
		articleCategory.SaveAll(&dataList)
	}

}
