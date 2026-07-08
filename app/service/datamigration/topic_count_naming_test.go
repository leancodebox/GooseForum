package datamigration

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestMigrateTopicCountNamingRenamesLegacyUserStatisticsColumn(t *testing.T) {
	conn := openTopicCountMigrationDB(t)
	mustExecTopicCountMigration(t, conn, `CREATE TABLE user_statistics (
		user_id integer primary key,
		article_count integer not null default 0,
		reply_count integer not null default 0
	)`)
	mustExecTopicCountMigration(t, conn, `INSERT INTO user_statistics (user_id, article_count, reply_count) VALUES (1, 7, 3)`)
	createDailyStatsForTopicCountMigration(t, conn)
	mustExecTopicCountMigration(t, conn, `INSERT INTO daily_stats (stat_date, stat_key, stat_value) VALUES ('2026-07-01', 'article_count', 5)`)

	result := MigrateTopicCountNamingWithDB(conn)
	if result.Failed != 0 {
		t.Fatalf("MigrateTopicCountNamingWithDB() failed = %d last=%s", result.Failed, result.LastFailed)
	}
	if conn.Migrator().HasColumn("user_statistics", "article_count") {
		t.Fatalf("legacy article_count column still exists")
	}
	if !conn.Migrator().HasColumn("user_statistics", "topic_count") {
		t.Fatalf("topic_count column was not created")
	}

	var topicCount int
	if err := conn.Table("user_statistics").Select("topic_count").Where("user_id = ?", 1).Scan(&topicCount).Error; err != nil {
		t.Fatalf("scan topic_count: %v", err)
	}
	if topicCount != 7 {
		t.Fatalf("topic_count = %d, want 7", topicCount)
	}
	assertDailyStatTopicCount(t, conn, "2026-07-01", 5)
	assertNoDailyStatArticleCount(t, conn)
}

func TestMigrateTopicCountNamingCopiesWhenBothColumnsExist(t *testing.T) {
	conn := openTopicCountMigrationDB(t)
	mustExecTopicCountMigration(t, conn, `CREATE TABLE user_statistics (
		user_id integer primary key,
		article_count integer not null default 0,
		topic_count integer not null default 0,
		reply_count integer not null default 0
	)`)
	mustExecTopicCountMigration(t, conn, `INSERT INTO user_statistics (user_id, article_count, topic_count, reply_count) VALUES (1, 9, 0, 4)`)
	createDailyStatsForTopicCountMigration(t, conn)
	mustExecTopicCountMigration(t, conn, `INSERT INTO daily_stats (stat_date, stat_key, stat_value) VALUES ('2026-07-01', 'article_count', 5)`)
	mustExecTopicCountMigration(t, conn, `INSERT INTO daily_stats (stat_date, stat_key, stat_value) VALUES ('2026-07-01', 'topic_count', 2)`)

	result := MigrateTopicCountNamingWithDB(conn)
	if result.Failed != 0 {
		t.Fatalf("MigrateTopicCountNamingWithDB() failed = %d last=%s", result.Failed, result.LastFailed)
	}
	if conn.Migrator().HasColumn("user_statistics", "article_count") {
		t.Fatalf("legacy article_count column still exists")
	}

	var topicCount int
	if err := conn.Table("user_statistics").Select("topic_count").Where("user_id = ?", 1).Scan(&topicCount).Error; err != nil {
		t.Fatalf("scan topic_count: %v", err)
	}
	if topicCount != 9 {
		t.Fatalf("topic_count = %d, want copied article_count 9", topicCount)
	}
	assertDailyStatTopicCount(t, conn, "2026-07-01", 7)
	assertNoDailyStatArticleCount(t, conn)
}

func TestMigrateTopicCountNamingSkipsCurrentSchema(t *testing.T) {
	conn := openTopicCountMigrationDB(t)
	mustExecTopicCountMigration(t, conn, `CREATE TABLE user_statistics (
		user_id integer primary key,
		topic_count integer not null default 0,
		reply_count integer not null default 0
	)`)
	mustExecTopicCountMigration(t, conn, `INSERT INTO user_statistics (user_id, topic_count, reply_count) VALUES (1, 11, 4)`)
	createDailyStatsForTopicCountMigration(t, conn)
	mustExecTopicCountMigration(t, conn, `INSERT INTO daily_stats (stat_date, stat_key, stat_value) VALUES ('2026-07-01', 'topic_count', 3)`)

	result := MigrateTopicCountNamingWithDB(conn)
	if result.Failed != 0 {
		t.Fatalf("MigrateTopicCountNamingWithDB() failed = %d last=%s", result.Failed, result.LastFailed)
	}

	var topicCount int
	if err := conn.Table("user_statistics").Select("topic_count").Where("user_id = ?", 1).Scan(&topicCount).Error; err != nil {
		t.Fatalf("scan topic_count: %v", err)
	}
	if topicCount != 11 {
		t.Fatalf("topic_count = %d, want 11", topicCount)
	}
	assertDailyStatTopicCount(t, conn, "2026-07-01", 3)
	assertNoDailyStatArticleCount(t, conn)
}

func openTopicCountMigrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	conn, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	return conn
}

func createDailyStatsForTopicCountMigration(t *testing.T, conn *gorm.DB) {
	t.Helper()
	mustExecTopicCountMigration(t, conn, `CREATE TABLE daily_stats (
		stat_date date not null,
		stat_key varchar(128) not null,
		stat_value bigint not null default 0,
		primary key (stat_date, stat_key)
	)`)
}

func mustExecTopicCountMigration(t *testing.T, conn *gorm.DB, sql string) {
	t.Helper()
	if err := conn.Exec(sql).Error; err != nil {
		t.Fatalf("exec %q: %v", sql, err)
	}
}

func assertDailyStatTopicCount(t *testing.T, conn *gorm.DB, date string, want int64) {
	t.Helper()
	var got int64
	if err := conn.Table("daily_stats").
		Select("stat_value").
		Where("date(stat_date) = ? AND stat_key = ?", date, "topic_count").
		Scan(&got).Error; err != nil {
		t.Fatalf("scan daily topic_count: %v", err)
	}
	if got != want {
		t.Fatalf("daily topic_count = %d, want %d", got, want)
	}
}

func assertNoDailyStatArticleCount(t *testing.T, conn *gorm.DB) {
	t.Helper()
	var count int64
	if err := conn.Table("daily_stats").Where("stat_key = ?", "article_count").Count(&count).Error; err != nil {
		t.Fatalf("count article_count daily stats: %v", err)
	}
	if count != 0 {
		t.Fatalf("article_count daily stat rows = %d, want 0", count)
	}
}
