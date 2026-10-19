package datamigration

import (
	"encoding/json"
	"errors"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func MigrateRegistrationCooldown() error {
	return MigrateRegistrationCooldownWithDB(dbconnect.Connect())
}

func MigrateRegistrationCooldownWithDB(db *gorm.DB) error {
	var row pageConfig.Entity
	if err := db.Where("page_type = ?", pageConfig.TrustSettings).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	config, err := pageConfig.DecodeTrustSettings([]byte(row.Config))
	if err != nil {
		return err
	}
	var header struct {
		SchemaVersion int `json:"schemaVersion"`
	}
	if err := json.Unmarshal([]byte(row.Config), &header); err != nil {
		return err
	}
	if header.SchemaVersion == 3 {
		return nil
	}
	// Keep the previous policy for administrators to inspect differing level values.
	previous := pageConfig.Entity{PageType: "trustSettingsBeforeSharedCooldown", Config: row.Config}
	if err := db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "page_type"}}, DoNothing: true}).Create(&previous).Error; err != nil {
		return err
	}
	raw, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return db.Model(&pageConfig.Entity{}).Where("id = ?", row.Id).Update("config", string(raw)).Error
}
