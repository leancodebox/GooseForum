package migration

import (
	"fmt"
	"log/slog"

	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/service/datamigration"
)

func runVersionedDataMigrations() error {
	currentVersion := pageConfig.GetMigrationVersion()
	if currentVersion >= pageConfig.AppMigrationVersion {
		return nil
	}

	slog.Info("app migration start", "currentVersion", currentVersion, "targetVersion", pageConfig.AppMigrationVersion)
	if currentVersion < 1 {
		datamigration.EnsureDefaultData()
		result := datamigration.RebuildReplyMarkdown()
		slog.Info("app migration rebuild reply markdown done", "processed", result.Processed, "skipped", result.Skipped, "failed", result.Failed)
		if result.Failed > 0 {
			slog.Error("app migration rebuild reply markdown has failures", "failed", result.Failed)
			return fmt.Errorf("rebuild reply markdown: %d failures", result.Failed)
		}
		if err := syncMigrationVersion(1); err != nil {
			return err
		}
		currentVersion = 1
	}
	if currentVersion < 2 {
		result := datamigration.BackfillReplySequence()
		slog.Info("app migration backfill reply sequence done", "articles", result.Articles, "replies", result.Replies, "skipped", result.Skipped, "failed", result.Failed)
		if result.Failed > 0 {
			slog.Error("app migration backfill reply sequence has failures", "failed", result.Failed, "lastFailed", result.LastFailed)
			return fmt.Errorf("backfill reply sequence: last failed topic %d", result.LastFailed)
		}
		if err := syncMigrationVersion(2); err != nil {
			return err
		}
		currentVersion = 2
	}
	if currentVersion < 3 {
		result := datamigration.BackfillArticleUserAction()
		slog.Info("app migration backfill article user action done", "processed", result.Processed, "skipped", result.Skipped, "failed", result.Failed)
		if result.Failed > 0 {
			slog.Error("app migration backfill article user action has failures", "failed", result.Failed)
			return fmt.Errorf("backfill article user action: %d failures", result.Failed)
		}
		if err := syncMigrationVersion(3); err != nil {
			return err
		}
		currentVersion = 3
	}
	if currentVersion < 4 {
		result := datamigration.MigrateSiteChromeContent()
		slog.Info("app migration site chrome content done", "migrated", result.Migrated, "failed", result.Failed)
		if result.Failed > 0 {
			slog.Error("app migration site chrome content has failures", "failed", result.Failed)
			return fmt.Errorf("migrate site chrome content: %d failures", result.Failed)
		}
		if err := syncMigrationVersion(4); err != nil {
			return err
		}
		currentVersion = 4
	}
	if currentVersion < 5 {
		result := datamigration.BackfillTopicPostModel()
		slog.Info("app migration topic post model done", "topics", result.Topics, "posts", result.Posts, "categories", result.Categories, "topicCategoryIndexes", result.TopicCategoryIndexes, "topicUserActions", result.TopicUserActions, "topicUserStats", result.TopicUserStats, "mappings", result.Mappings, "notifications", result.Notifications, "reportsChecked", result.ReportsChecked, "reportsMissing", result.ReportsMissing, "moderationLogs", result.ModerationLogs, "moderationLogsMissing", result.ModerationLogsMissing, "skipped", result.Skipped, "failed", result.Failed, "lastFailed", result.LastFailed)
		if result.Failed > 0 {
			slog.Error("app migration topic post model has failures", "failed", result.Failed, "lastFailed", result.LastFailed)
			return fmt.Errorf("backfill topic post model: %s", result.LastFailed)
		}
		if err := syncMigrationVersion(5); err != nil {
			return err
		}
		currentVersion = 5
	}
	if currentVersion < 6 {
		result := datamigration.BackfillModerationLogsTopicPost()
		slog.Info("app migration moderation log topic post migration done", "moderationLogs", result.ModerationLogs, "moderationLogsMissing", result.ModerationLogsMissing, "failed", result.Failed, "lastFailed", result.LastFailed)
		if result.Failed > 0 {
			slog.Error("app migration moderation log topic post migration has failures", "failed", result.Failed, "lastFailed", result.LastFailed)
			return fmt.Errorf("backfill moderation log topic post: %s", result.LastFailed)
		}
		if err := syncMigrationVersion(6); err != nil {
			return err
		}
		currentVersion = 6
	}
	if currentVersion < 7 {
		result := datamigration.BackfillFileUsagesTopicPost()
		slog.Info("app migration file usage topic post migration done", "fileUsages", result.FileUsages, "fileUsagesMissing", result.FileUsagesMissing, "failed", result.Failed, "lastFailed", result.LastFailed)
		if result.Failed > 0 {
			slog.Error("app migration file usage topic post migration has failures", "failed", result.Failed, "lastFailed", result.LastFailed)
			return fmt.Errorf("backfill file usages topic post: %s", result.LastFailed)
		}
		if err := syncMigrationVersion(7); err != nil {
			return err
		}
		currentVersion = 7
	}
	if currentVersion < 8 {
		result := datamigration.MigrateTopicCountNaming()
		slog.Info("app migration topic count naming done", "userStatisticsMigrated", result.UserStatisticsMigrated, "dailyStatsMigrated", result.DailyStatsMigrated, "failed", result.Failed, "lastFailed", result.LastFailed)
		if result.Failed > 0 {
			slog.Error("app migration topic count naming has failures", "failed", result.Failed, "lastFailed", result.LastFailed)
			return fmt.Errorf("migrate topic count naming: %s", result.LastFailed)
		}
		if err := syncMigrationVersion(8); err != nil {
			return err
		}
		currentVersion = 8
	}
	if currentVersion < 9 {
		result := datamigration.MigrateTopicSearchIndex()
		slog.Info("app migration topic search index done", "skipped", result.Skipped, "rebuilt", result.Rebuilt, "processed", result.ProcessedCount, "failedCount", result.FailedCount, "legacyIndexDeleteTried", result.LegacyIndexDeleteTried, "legacyIndexDeleted", result.LegacyIndexDeleted, "failed", result.Failed, "lastFailed", result.LastFailed)
		if result.Failed > 0 || result.FailedCount > 0 {
			slog.Error("app migration topic search index has failures", "failed", result.Failed, "failedCount", result.FailedCount, "lastFailed", result.LastFailed)
			return fmt.Errorf("migrate topic search index: %s", result.LastFailed)
		}
		if err := syncMigrationVersion(9); err != nil {
			return err
		}
		currentVersion = 9
	}
	if currentVersion < 10 {
		result := datamigration.DropReportLegacyColumns()
		slog.Info("app migration report legacy columns done", "articleIDColumnDropped", result.ArticleIDColumnDropped, "statusArticleIndexDrop", result.StatusArticleIndexDrop, "articleIndexDrop", result.ArticleIndexDrop, "failed", result.Failed, "lastFailed", result.LastFailed)
		if result.Failed > 0 {
			slog.Error("app migration report legacy columns has failures", "failed", result.Failed, "lastFailed", result.LastFailed)
			return fmt.Errorf("drop report legacy columns: %s", result.LastFailed)
		}
		if err := syncMigrationVersion(10); err != nil {
			return err
		}
		currentVersion = 10
	}
	if currentVersion < 11 {
		result := datamigration.MigratePointsRecordAction()
		slog.Info("app migration points record action done", "backfilled", result.Backfilled, "changeReasonColumnDropped", result.ChangeReasonColumnDropped, "failed", result.Failed, "lastFailed", result.LastFailed)
		if result.Failed > 0 {
			slog.Error("app migration points record action has failures", "failed", result.Failed, "lastFailed", result.LastFailed)
			return fmt.Errorf("migrate points record action: %s", result.LastFailed)
		}
		if err := syncMigrationVersion(11); err != nil {
			return err
		}
		currentVersion = 11
	}
	if currentVersion < 12 {
		result := datamigration.RebuildPostMarkdown()
		slog.Info("app migration rebuild post markdown done", "processed", result.Processed, "failed", result.Failed, "lastFailed", result.LastFailed)
		if result.Failed > 0 {
			slog.Error("app migration rebuild post markdown has failures", "failed", result.Failed, "lastFailed", result.LastFailed)
			return fmt.Errorf("rebuild post markdown: %s", result.LastFailed)
		}
		if err := syncMigrationVersion(12); err != nil {
			return err
		}
		currentVersion = 12
	}
	if currentVersion < 13 {
		result := datamigration.BackfillAccessControlDefaults()
		slog.Info("app migration access control defaults done", "groups", result.Groups, "categories", result.Categories, "grants", result.Grants, "failed", result.Failed, "lastFailed", result.LastFailed)
		if result.Failed > 0 {
			slog.Error("app migration access control defaults has failures", "failed", result.Failed, "lastFailed", result.LastFailed)
			return fmt.Errorf("backfill access control defaults: %s", result.LastFailed)
		}
		if err := syncMigrationVersion(13); err != nil {
			return err
		}
		currentVersion = 13
	}
	if currentVersion < 14 {
		result := datamigration.BackfillUserActivityTopicIDs()
		slog.Info("app migration user activity topic ids done", "topics", result.Topics, "posts", result.Posts, "missing", result.Missing, "failed", result.Failed, "lastFailed", result.LastFailed)
		if result.Failed > 0 {
			slog.Error("app migration user activity topic ids has failures", "failed", result.Failed, "lastFailed", result.LastFailed)
			return fmt.Errorf("backfill user activity topic ids: %s", result.LastFailed)
		}
		if err := syncMigrationVersion(14); err != nil {
			return err
		}
		currentVersion = 14
	}
	if currentVersion < 15 {
		result := datamigration.BackfillNotificationTopicIDs()
		slog.Info("app migration notification topic ids done", "updated", result.Updated, "failed", result.Failed, "lastFailed", result.LastFailed)
		if result.Failed > 0 {
			slog.Error("app migration notification topic ids has failures", "failed", result.Failed, "lastFailed", result.LastFailed)
			return fmt.Errorf("backfill notification topic ids: %s", result.LastFailed)
		}
		if err := syncMigrationVersion(15); err != nil {
			return err
		}
		currentVersion = 15
	}
	if currentVersion < 16 {
		result := datamigration.BackfillTopicMainCategory()
		slog.Info("app migration topic main category done", "updated", result.Updated, "orphaned", result.Orphaned, "failed", result.Failed, "lastFailed", result.LastFailed)
		if result.Failed > 0 {
			slog.Error("app migration topic main category has failures", "failed", result.Failed, "lastFailed", result.LastFailed)
			return fmt.Errorf("backfill topic main category: %s", result.LastFailed)
		}
		searchResult := datamigration.RebuildTopicMainCategorySearchIndex()
		slog.Info("app migration topic main category search index done", "skipped", searchResult.Skipped, "rebuilt", searchResult.Rebuilt, "processed", searchResult.ProcessedCount, "failedCount", searchResult.FailedCount, "failed", searchResult.Failed, "lastFailed", searchResult.LastFailed)
		if searchResult.Failed > 0 || searchResult.FailedCount > 0 {
			slog.Error("app migration topic main category search index has failures", "failed", searchResult.Failed, "failedCount", searchResult.FailedCount, "lastFailed", searchResult.LastFailed)
			return fmt.Errorf("rebuild topic main category search index: %s", searchResult.LastFailed)
		}
		if err := syncMigrationVersion(16); err != nil {
			return err
		}
		currentVersion = 16
	}
	if currentVersion < 17 {
		result := datamigration.EnforceSingleRestrictedTopicCategory()
		slog.Info("app migration single restricted topic category done", "updated", result.Updated, "removedCategoryLinks", result.RemovedCategoryLinks, "failed", result.Failed, "lastFailed", result.LastFailed)
		if result.Failed > 0 {
			slog.Error("app migration single restricted topic category has failures", "failed", result.Failed, "lastFailed", result.LastFailed)
			return fmt.Errorf("enforce single restricted topic category: %s", result.LastFailed)
		}
		// Always retry the search rebuild while v17 is pending. The database repair
		// may have committed on a previous start whose search rebuild then failed.
		searchResult := datamigration.RebuildTopicMainCategorySearchIndex()
		slog.Info("app migration single restricted topic category search index done", "skipped", searchResult.Skipped, "rebuilt", searchResult.Rebuilt, "processed", searchResult.ProcessedCount, "failedCount", searchResult.FailedCount, "failed", searchResult.Failed, "lastFailed", searchResult.LastFailed)
		if searchResult.Failed > 0 || searchResult.FailedCount > 0 {
			slog.Error("app migration single restricted topic category search index has failures", "failed", searchResult.Failed, "failedCount", searchResult.FailedCount, "lastFailed", searchResult.LastFailed)
			return fmt.Errorf("rebuild single restricted topic category search index: %s", searchResult.LastFailed)
		}
		if err := syncMigrationVersion(17); err != nil {
			return err
		}
		currentVersion = 17
	}
	slog.Info("app migration end", "version", currentVersion)
	return nil
}

func syncMigrationVersion(version uint32) error {
	if err := pageConfig.SyncMigrationVersion(version); err != nil {
		return fmt.Errorf("persist app migration version %d: %w", version, err)
	}
	return nil
}
