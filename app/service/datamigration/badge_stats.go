package datamigration

import (
	"errors"
	"sort"

	"github.com/leancodebox/GooseForum/app/models/forum/badgeStats"
	"github.com/leancodebox/GooseForum/app/models/forum/badges"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BackfillBadgeStatsWithDB is only called by the versioned startup migration.
// Runtime requests read stored statistics and never rebuild them.
func BackfillBadgeStatsWithDB(conn *gorm.DB, systemCodes []string) error {
	if conn == nil {
		return errors.New("database is required")
	}
	return conn.Transaction(func(tx *gorm.DB) error {
		byCode := make(map[string]badgeStats.Entity)
		addCodes := func(codes []string) {
			for _, code := range codes {
				if code != "" {
					byCode[code] = badgeStats.Entity{BadgeCode: code}
				}
			}
		}
		addCodes(systemCodes)
		var codes []string
		if err := tx.Model(&badges.Entity{}).Pluck("code", &codes).Error; err != nil {
			return err
		}
		addCodes(codes)
		var existing []badgeStats.Entity
		if err := tx.Find(&existing).Error; err != nil {
			return err
		}
		for _, row := range existing {
			// History is permanent even if an older grant record was cleaned up.
			byCode[row.BadgeCode] = badgeStats.Entity{BadgeCode: row.BadgeCode, HasGrantHistory: row.HasGrantHistory}
		}
		var counts []struct {
			BadgeCode    string
			GrantedCount int64
		}
		if err := tx.Table("user_badges AS grants").
			Joins("LEFT JOIN users AS recipient ON recipient.id = grants.user_id").
			Select("grants.badge_code, SUM(CASE WHEN grants.revoked_at IS NULL AND recipient.id IS NOT NULL AND recipient.deleted_at IS NULL THEN 1 ELSE 0 END) AS granted_count").
			Group("grants.badge_code").Scan(&counts).Error; err != nil {
			return err
		}
		for _, count := range counts {
			byCode[count.BadgeCode] = badgeStats.Entity{BadgeCode: count.BadgeCode, GrantedCount: count.GrantedCount, HasGrantHistory: true}
		}
		rows := make([]badgeStats.Entity, 0, len(byCode))
		for _, row := range byCode {
			rows = append(rows, row)
		}
		sort.Slice(rows, func(i, j int) bool { return rows[i].BadgeCode < rows[j].BadgeCode })
		if len(rows) == 0 {
			return nil
		}
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "badge_code"}},
			DoUpdates: clause.AssignmentColumns([]string{"granted_count", "has_grant_history"}),
		}).CreateInBatches(&rows, 100).Error
	})
}
