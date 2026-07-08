package datamigration

import (
	db "github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TopicCountNamingResult struct {
	UserStatisticsMigrated bool
	DailyStatsMigrated     int64
	Failed                 int
	LastFailed             string
}

func MigrateTopicCountNaming() TopicCountNamingResult {
	return MigrateTopicCountNamingWithDB(db.Connect())
}

func MigrateTopicCountNamingWithDB(conn *gorm.DB) TopicCountNamingResult {
	result := TopicCountNamingResult{}
	migrateUserStatisticsTopicCount(conn, &result)
	migrateDailyStatsTopicCount(conn, &result)
	return result
}

func migrateUserStatisticsTopicCount(conn *gorm.DB, result *TopicCountNamingResult) {
	if !conn.Migrator().HasTable("user_statistics") {
		return
	}

	hasArticleCount := conn.Migrator().HasColumn("user_statistics", "article_count")
	hasTopicCount := conn.Migrator().HasColumn("user_statistics", "topic_count")
	if !hasArticleCount {
		return
	}
	if !hasTopicCount {
		if err := conn.Migrator().RenameColumn("user_statistics", "article_count", "topic_count"); err != nil {
			failTopicCountMigration(result, "user_statistics_rename", err)
			return
		}
		result.UserStatisticsMigrated = true
		return
	}

	if err := conn.Exec("UPDATE user_statistics SET topic_count = article_count").Error; err != nil {
		failTopicCountMigration(result, "user_statistics_copy", err)
		return
	}
	if err := dropLegacyTopicCountColumn(conn); err != nil {
		failTopicCountMigration(result, "user_statistics_drop_legacy", err)
		return
	}
	result.UserStatisticsMigrated = true
}

type legacyDailyTopicCount struct {
	StatDate  string
	StatValue int64
}

func migrateDailyStatsTopicCount(conn *gorm.DB, result *TopicCountNamingResult) {
	if !conn.Migrator().HasTable("daily_stats") {
		return
	}

	var rows []legacyDailyTopicCount
	if err := conn.Table("daily_stats").
		Select("date(stat_date) AS stat_date, stat_value").
		Where("stat_key = ?", "article_count").
		Find(&rows).Error; err != nil {
		failTopicCountMigration(result, "daily_stats_scan", err)
		return
	}
	for _, row := range rows {
		if err := conn.Table("daily_stats").Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "stat_date"}, {Name: "stat_key"}},
			DoUpdates: clause.Assignments(map[string]any{
				"stat_value": gorm.Expr("stat_value + ?", row.StatValue),
			}),
		}).Create(map[string]any{
			"stat_date":  row.StatDate,
			"stat_key":   "topic_count",
			"stat_value": row.StatValue,
		}).Error; err != nil {
			failTopicCountMigration(result, "daily_stats_upsert", err)
			return
		}
		result.DailyStatsMigrated++
	}
	if len(rows) == 0 {
		return
	}
	if err := conn.Exec("DELETE FROM daily_stats WHERE stat_key = ?", "article_count").Error; err != nil {
		failTopicCountMigration(result, "daily_stats_delete_legacy", err)
		return
	}
}

func dropLegacyTopicCountColumn(conn *gorm.DB) error {
	return conn.Exec("ALTER TABLE user_statistics DROP COLUMN article_count").Error
}

func failTopicCountMigration(result *TopicCountNamingResult, step string, err error) {
	result.Failed++
	result.LastFailed = step + ": " + err.Error()
}
