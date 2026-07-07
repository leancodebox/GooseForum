package datamigration

import (
	"log/slog"
	"time"

	db "github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/topicUserAction"
	"gorm.io/gorm/clause"
)

type ArticleUserActionResult struct {
	Processed int
	Skipped   int
	Failed    int
}

func BackfillArticleUserAction() ArticleUserActionResult {
	conn := db.Connect()
	result := ArticleUserActionResult{}
	if err := conn.AutoMigrate(&topicUserAction.Entity{}); err != nil {
		result.Failed++
		slog.Error("article user action table migration failed", "err", err)
		return result
	}
	sources := []struct {
		table string
		field string
	}{
		{table: "article_like", field: "liked_at"},
		{table: "article_bookmark", field: "bookmarked_at"},
		{table: "article_watch", field: "watched_at"},
	}

	for _, source := range sources {
		if !conn.Migrator().HasTable(source.table) {
			continue
		}
		var rows []struct {
			UserId    uint64
			ArticleId uint64
			UpdatedAt time.Time
		}
		if err := conn.Table(source.table).Where("status = ?", 1).Find(&rows).Error; err != nil {
			result.Failed++
			slog.Error("article user action source scan failed", "table", source.table, "err", err)
			continue
		}
		for _, row := range rows {
			if row.UserId == 0 || row.ArticleId == 0 {
				result.Skipped++
				continue
			}
			actedAt := row.UpdatedAt
			if actedAt.IsZero() {
				actedAt = time.Now()
			}
			rowData := map[string]any{
				"user_id":    row.UserId,
				"topic_id":   row.ArticleId,
				source.field: actedAt,
				"created_at": actedAt,
				"updated_at": actedAt,
			}
			if err := conn.Table("topic_user_action").Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "user_id"}, {Name: "topic_id"}},
				DoUpdates: clause.Assignments(map[string]any{
					source.field: actedAt,
				}),
			}).Create(rowData).Error; err != nil {
				result.Failed++
				slog.Error("topic user action backfill failed", "table", source.table, "userId", row.UserId, "articleId", row.ArticleId, "err", err)
				continue
			}
			result.Processed++
		}
	}

	return result
}
