package cmd

import (
	"database/sql"
	_ "embed"
	"fmt"
	"github.com/leancodebox/goose/preferences"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

//go:embed dbinit/init.sql
var dataAllSql string

func init() {
	appendCommand(&cobra.Command{
		Use:   "bbs:init:db",
		Short: "bbs数据库初始化",
		Run:   runBbsinitDB,
		// Args:  cobra.ExactArgs(1), // 只允许且必须传 1 个参数
	})
}

func runBbsinitDB(_ *cobra.Command, _ []string) {
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
