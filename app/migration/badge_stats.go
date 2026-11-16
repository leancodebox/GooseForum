package migration

import (
	"fmt"

	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/service/datamigration"
	"gorm.io/gorm"
)

func migrateBadgeStatsWithDB(conn *gorm.DB, systemCodes []string) error {
	return conn.Transaction(func(tx *gorm.DB) error {
		version, err := pageConfig.GetMigrationVersionWithDB(tx)
		if err != nil || version >= 31 {
			return err
		}
		if err := datamigration.BackfillBadgeStatsWithDB(tx, systemCodes); err != nil {
			return fmt.Errorf("backfill badge statistics: %w", err)
		}
		if err := pageConfig.SyncMigrationVersionWithDB(tx, 31); err != nil {
			return fmt.Errorf("persist badge statistics migration version: %w", err)
		}
		return nil
	})
}
