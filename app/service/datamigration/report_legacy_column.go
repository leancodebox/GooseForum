package datamigration

import (
	"fmt"

	db "github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"gorm.io/gorm"
)

type ReportLegacyColumnResult struct {
	ArticleIDColumnDropped bool
	StatusArticleIndexDrop bool
	ArticleIndexDrop       bool
	Failed                 int
	LastFailed             string
}

func DropReportLegacyColumns() ReportLegacyColumnResult {
	return DropReportLegacyColumnsWithDB(db.Connect())
}

func DropReportLegacyColumnsWithDB(conn *gorm.DB) ReportLegacyColumnResult {
	result := ReportLegacyColumnResult{}
	if !conn.Migrator().HasTable("reports") {
		return result
	}

	if conn.Migrator().HasIndex("reports", "idx_reports_status_article_id") {
		if err := dropReportLegacyIndex(conn, "idx_reports_status_article_id"); err != nil {
			failReportLegacyColumnMigration(&result, "drop_status_article_index", err)
			return result
		}
		result.StatusArticleIndexDrop = true
	}
	if conn.Migrator().HasIndex("reports", "idx_reports_article_id") {
		if err := dropReportLegacyIndex(conn, "idx_reports_article_id"); err != nil {
			failReportLegacyColumnMigration(&result, "drop_article_index", err)
			return result
		}
		result.ArticleIndexDrop = true
	}
	if !conn.Migrator().HasColumn("reports", "article_id") {
		return result
	}
	if err := conn.Exec("ALTER TABLE reports DROP COLUMN article_id").Error; err != nil {
		failReportLegacyColumnMigration(&result, "drop_article_id_column", err)
		return result
	}
	result.ArticleIDColumnDropped = true
	return result
}

func dropReportLegacyIndex(conn *gorm.DB, indexName string) error {
	switch conn.Dialector.Name() {
	case "mysql":
		return conn.Exec("DROP INDEX " + indexName + " ON reports").Error
	default:
		return conn.Exec("DROP INDEX IF EXISTS " + indexName).Error
	}
}

func failReportLegacyColumnMigration(result *ReportLegacyColumnResult, step string, err error) {
	result.Failed++
	result.LastFailed = fmt.Sprintf("%s: %s", step, err.Error())
}
