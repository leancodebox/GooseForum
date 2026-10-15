package datamigration

import (
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/trustlevel"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroups"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func InitializeTrustLevels() error { return InitializeTrustLevelsWithDB(dbconnect.Connect()) }
func InitializeTrustLevelsWithDB(db *gorm.DB) error {
	for _, d := range trustlevel.Definitions() {
		if d.SystemKey == "" {
			continue
		}
		g := accessGroups.Entity{Name: d.Name + "及以上", SystemKey: &d.SystemKey, JoinMode: accessGroups.JoinModeSystem, Status: accessGroups.StatusEnabled}
		if err := db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "system_key"}}, DoNothing: true}).Create(&g).Error; err != nil {
			return err
		}
	}
	// Runs before serving registrations, only in the versioned migration.
	return db.Model(&users.EntityComplete{}).Where("trust_level = ?", trustlevel.Newcomer).Update("trust_level", trustlevel.Member).Error
}
