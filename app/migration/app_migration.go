package migration

import (
	"log/slog"

	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/service/datamigration"
)

func runVersionedDataMigrations() {
	currentVersion := pageConfig.GetMigrationVersion()
	if currentVersion >= pageConfig.AppMigrationVersion {
		return
	}

	slog.Info("app migration start", "currentVersion", currentVersion, "targetVersion", pageConfig.AppMigrationVersion)
	if currentVersion < 1 {
		datamigration.EnsureDefaultData()
		result := datamigration.RebuildReplyMarkdown()
		slog.Info("app migration rebuild reply markdown done", "processed", result.Processed, "skipped", result.Skipped, "failed", result.Failed)
		if result.Failed > 0 {
			slog.Error("app migration rebuild reply markdown has failures", "failed", result.Failed)
			return
		}
		pageConfig.SyncMigrationVersion(1)
		currentVersion = 1
	}
	if currentVersion < 2 {
		result := datamigration.BackfillReplySequence()
		slog.Info("app migration backfill reply sequence done", "articles", result.Articles, "replies", result.Replies, "skipped", result.Skipped, "failed", result.Failed)
		if result.Failed > 0 {
			slog.Error("app migration backfill reply sequence has failures", "failed", result.Failed, "lastFailed", result.LastFailed)
			return
		}
		pageConfig.SyncMigrationVersion(2)
		currentVersion = 2
	}
	if currentVersion < 3 {
		result := datamigration.BackfillArticleUserAction()
		slog.Info("app migration backfill article user action done", "processed", result.Processed, "skipped", result.Skipped, "failed", result.Failed)
		if result.Failed > 0 {
			slog.Error("app migration backfill article user action has failures", "failed", result.Failed)
			return
		}
		pageConfig.SyncMigrationVersion(3)
		currentVersion = 3
	}
	if currentVersion < 4 {
		result := datamigration.MigrateSiteChromeContent()
		slog.Info("app migration site chrome content done", "migrated", result.Migrated, "failed", result.Failed)
		if result.Failed > 0 {
			slog.Error("app migration site chrome content has failures", "failed", result.Failed)
			return
		}
		pageConfig.SyncMigrationVersion(4)
		currentVersion = 4
	}
	if currentVersion < 5 {
		result := datamigration.BackfillTopicPostModel()
		slog.Info("app migration topic post model done", "topics", result.Topics, "posts", result.Posts, "categories", result.Categories, "topicCategoryIndexes", result.TopicCategoryIndexes, "topicUserActions", result.TopicUserActions, "topicUserStats", result.TopicUserStats, "mappings", result.Mappings, "notifications", result.Notifications, "reportsChecked", result.ReportsChecked, "reportsMissing", result.ReportsMissing, "moderationLogs", result.ModerationLogs, "moderationLogsMissing", result.ModerationLogsMissing, "skipped", result.Skipped, "failed", result.Failed, "lastFailed", result.LastFailed)
		if result.Failed > 0 {
			slog.Error("app migration topic post model has failures", "failed", result.Failed, "lastFailed", result.LastFailed)
			return
		}
		pageConfig.SyncMigrationVersion(5)
		currentVersion = 5
	}
	if currentVersion < 6 {
		result := datamigration.BackfillModerationLogsTopicPost()
		slog.Info("app migration moderation log topic post migration done", "moderationLogs", result.ModerationLogs, "moderationLogsMissing", result.ModerationLogsMissing, "failed", result.Failed, "lastFailed", result.LastFailed)
		if result.Failed > 0 {
			slog.Error("app migration moderation log topic post migration has failures", "failed", result.Failed, "lastFailed", result.LastFailed)
			return
		}
		pageConfig.SyncMigrationVersion(6)
		currentVersion = 6
	}
	slog.Info("app migration end", "version", currentVersion)
}
