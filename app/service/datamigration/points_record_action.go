package datamigration

import (
	"fmt"

	db "github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"gorm.io/gorm"
)

type PointsRecordActionResult struct {
	Backfilled                int64
	ChangeReasonColumnDropped bool
	Failed                    int
	LastFailed                string
}

func MigratePointsRecordAction() PointsRecordActionResult {
	return MigratePointsRecordActionWithDB(db.Connect())
}

func MigratePointsRecordActionWithDB(conn *gorm.DB) PointsRecordActionResult {
	result := PointsRecordActionResult{}
	if !conn.Migrator().HasTable("points_record") {
		return result
	}
	if !conn.Migrator().HasColumn("points_record", "action") {
		return result
	}
	if !conn.Migrator().HasColumn("points_record", "change_reason") {
		return result
	}

	tx := conn.Table("points_record").
		Where("action = '' OR action IS NULL").
		Updates(map[string]any{
			"action": gorm.Expr(`CASE
				WHEN change_reason = ? THEN ?
				ELSE ?
			END`, "初始化", "init", "unknown"),
		})
	if tx.Error != nil {
		failPointsRecordActionMigration(&result, "backfill_action", tx.Error)
		return result
	}
	result.Backfilled = tx.RowsAffected
	if err := conn.Exec("ALTER TABLE points_record DROP COLUMN change_reason").Error; err != nil {
		failPointsRecordActionMigration(&result, "drop_change_reason_column", err)
		return result
	}
	result.ChangeReasonColumnDropped = true
	return result
}

func failPointsRecordActionMigration(result *PointsRecordActionResult, step string, err error) {
	result.Failed++
	result.LastFailed = fmt.Sprintf("%s: %s", step, err.Error())
}
