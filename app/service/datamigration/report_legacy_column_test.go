package datamigration

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestDropReportLegacyColumnsDropsArticleIDAndIndex(t *testing.T) {
	conn := openReportLegacyColumnDB(t)
	mustExecReportLegacyColumn(t, conn, `CREATE TABLE reports (
		id integer primary key,
		target_type text not null default '',
		target_id integer not null default 0,
		article_id integer not null default 0,
		topic_id integer not null default 0,
		reporter_id integer not null default 0,
		status text not null default 'open'
	)`)
	mustExecReportLegacyColumn(t, conn, `CREATE INDEX idx_reports_status_article_id ON reports(status, article_id, id)`)
	mustExecReportLegacyColumn(t, conn, `CREATE INDEX idx_reports_article_id ON reports(article_id)`)
	mustExecReportLegacyColumn(t, conn, `INSERT INTO reports (id, target_type, target_id, article_id, topic_id, reporter_id, status) VALUES (1, 'post', 88, 10, 10, 3, 'open')`)

	result := DropReportLegacyColumnsWithDB(conn)
	if result.Failed != 0 {
		t.Fatalf("DropReportLegacyColumnsWithDB() failed = %d last=%s", result.Failed, result.LastFailed)
	}
	if !result.StatusArticleIndexDrop || !result.ArticleIndexDrop || !result.ArticleIDColumnDropped {
		t.Fatalf("unexpected migration result: %#v", result)
	}
	if conn.Migrator().HasIndex("reports", "idx_reports_status_article_id") {
		t.Fatalf("legacy report status/article index still exists")
	}
	if conn.Migrator().HasIndex("reports", "idx_reports_article_id") {
		t.Fatalf("legacy report article index still exists")
	}
	if conn.Migrator().HasColumn("reports", "article_id") {
		t.Fatalf("legacy report article_id column still exists")
	}

	var topicID uint64
	if err := conn.Table("reports").Select("topic_id").Where("id = ?", 1).Scan(&topicID).Error; err != nil {
		t.Fatalf("scan topic_id: %v", err)
	}
	if topicID != 10 {
		t.Fatalf("topic_id = %d, want 10", topicID)
	}
}

func TestDropReportLegacyColumnsSkipsCurrentSchema(t *testing.T) {
	conn := openReportLegacyColumnDB(t)
	mustExecReportLegacyColumn(t, conn, `CREATE TABLE reports (
		id integer primary key,
		target_type text not null default '',
		target_id integer not null default 0,
		topic_id integer not null default 0,
		reporter_id integer not null default 0,
		status text not null default 'open'
	)`)

	result := DropReportLegacyColumnsWithDB(conn)
	if result.Failed != 0 {
		t.Fatalf("DropReportLegacyColumnsWithDB() failed = %d last=%s", result.Failed, result.LastFailed)
	}
	if result.StatusArticleIndexDrop || result.ArticleIndexDrop || result.ArticleIDColumnDropped {
		t.Fatalf("current schema should not be changed: %#v", result)
	}
}

func openReportLegacyColumnDB(t *testing.T) *gorm.DB {
	t.Helper()
	conn, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	return conn
}

func mustExecReportLegacyColumn(t *testing.T, conn *gorm.DB, sql string) {
	t.Helper()
	if err := conn.Exec(sql).Error; err != nil {
		t.Fatalf("exec %q: %v", sql, err)
	}
}
