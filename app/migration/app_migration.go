package migration

import (
	"log/slog"

	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/service/datamigration"
)

func RunAppMigrations() {
	currentVersion := pageConfig.GetMigrationVersion()
	if currentVersion >= pageConfig.AppMigrationVersion {
		return
	}

	slog.Info("app migration start", "currentVersion", currentVersion, "targetVersion", pageConfig.AppMigrationVersion)
	if currentVersion < 1 {
		result := datamigration.RebuildReplyMarkdown()
		slog.Info("app migration rebuild reply markdown done", "processed", result.Processed, "skipped", result.Skipped, "failed", result.Failed)
		if result.Failed > 0 {
			slog.Error("app migration rebuild reply markdown has failures", "failed", result.Failed)
			return
		}
		pageConfig.SyncMigrationVersion(1)
		currentVersion = 1
	}
	slog.Info("app migration end", "version", currentVersion)
}
