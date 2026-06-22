package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/migration"
	"github.com/spf13/cobra"
)

type sqliteIndex struct {
	Name    string
	TblName string
	SQL     string
}

func init() {
	cmd := &cobra.Command{
		Use:   "rebuild-sqlite-indexes",
		Short: "Drop default SQLite indexes and rerun default DB migration",
		RunE:  runRebuildSQLiteIndexes,
	}
	cmd.Flags().Bool("yes", false, "Confirm dropping all non-auto SQLite indexes")
	appendCommand(cmd)
}

func runRebuildSQLiteIndexes(cmd *cobra.Command, _ []string) error {
	totalStart := time.Now()
	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		return fmt.Errorf("refusing to drop indexes without --yes")
	}
	if !dbconnect.IsSqlite() {
		return fmt.Errorf("default database is not sqlite")
	}

	db := dbconnect.Connect()
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(1)

	var indexes []sqliteIndex
	stepStart := time.Now()
	if err = db.Raw(`
SELECT name, tbl_name, sql FROM sqlite_master
WHERE type = 'index'
AND name NOT LIKE 'sqlite_autoindex%'
`).Scan(&indexes).Error; err != nil {
		return err
	}
	beforeCount := len(indexes)
	fmt.Printf("default sqlite indexes before rebuild: %d\n", beforeCount)
	printStepDuration("list indexes", stepStart)

	stepStart = time.Now()
	for _, index := range indexes {
		fmt.Printf("drop index %s on %s\n", index.Name, index.TblName)
		if err := db.Exec("DROP INDEX IF EXISTS " + sqliteQuoteIdent(index.Name)).Error; err != nil {
			return fmt.Errorf("drop index %s: %w", index.Name, err)
		}
	}
	printStepDuration("drop indexes", stepStart)

	stepStart = time.Now()
	migration.M()
	printStepDuration("rerun migrations", stepStart)

	var afterCount int64
	stepStart = time.Now()
	if err = db.Raw(`
SELECT COUNT(*) FROM sqlite_master
WHERE type = 'index'
AND name NOT LIKE 'sqlite_autoindex%'
`).Scan(&afterCount).Error; err != nil {
		return err
	}
	printStepDuration("count rebuilt indexes", stepStart)

	stepStart = time.Now()
	if err = db.Exec("VACUUM").Error; err != nil {
		return fmt.Errorf("vacuum default sqlite db: %w", err)
	}
	printStepDuration("vacuum", stepStart)

	fmt.Printf("default sqlite indexes after rebuild: %d\n", afterCount)
	fmt.Printf("rebuilt default sqlite indexes, dropped %d indexes\n", beforeCount)
	fmt.Printf("total duration: %s\n", time.Since(totalStart).Round(time.Millisecond))
	return nil
}

func sqliteQuoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func printStepDuration(name string, start time.Time) {
	fmt.Printf("%s duration: %s\n", name, time.Since(start).Round(time.Millisecond))
}
