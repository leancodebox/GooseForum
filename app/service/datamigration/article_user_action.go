package datamigration

import (
	"log/slog"
	"time"

	db "github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/articleUserAction"
)

type ArticleUserActionResult struct {
	Processed int
	Skipped   int
	Failed    int
}

func BackfillArticleUserAction() ArticleUserActionResult {
	conn := db.Connect()
	result := ArticleUserActionResult{}
	if err := conn.AutoMigrate(&articleUserAction.Entity{}); err != nil {
		result.Failed++
		slog.Error("article user action table migration failed", "err", err)
		return result
	}
	sources := []struct {
		table string
		set   func(uint64, uint64, *time.Time) bool
	}{
		{table: "article_like", set: articleUserAction.SetLikedAt},
		{table: "article_bookmark", set: articleUserAction.SetBookmarkedAt},
		{table: "article_watch", set: articleUserAction.SetWatchedAt},
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
			if !source.set(row.UserId, row.ArticleId, &actedAt) {
				result.Failed++
				slog.Error("article user action backfill failed", "table", source.table, "userId", row.UserId, "articleId", row.ArticleId)
				continue
			}
			result.Processed++
		}
	}

	return result
}
