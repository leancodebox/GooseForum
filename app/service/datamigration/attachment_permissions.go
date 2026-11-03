package datamigration

import (
	"encoding/json"
	"errors"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"gorm.io/gorm"
)

func MigrateAttachmentPermissions() error {
	return MigrateAttachmentPermissionsWithDB(dbconnect.Connect())
}

func MigrateAttachmentPermissionsWithDB(db *gorm.DB) error {
	minimum := 100
	var posting pageConfig.Entity
	err := db.Where("page_type = ?", pageConfig.PostingSettings).First(&posting).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err == nil {
		var legacy struct {
			Attachments struct {
				MinTrustLevel *int `json:"minTrustLevel"`
			} `json:"attachments"`
		}
		if err := json.Unmarshal([]byte(posting.Config), &legacy); err != nil {
			return err
		}
		if legacy.Attachments.MinTrustLevel != nil {
			minimum = *legacy.Attachments.MinTrustLevel
		}
	}
	var row pageConfig.Entity
	err = db.Where("page_type = ?", pageConfig.TrustSettings).First(&row).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		config := pageConfig.DefaultTrustSettings()
		for i := range config.LevelLimits {
			config.LevelLimits[i].AllowAttachmentUpload = int(config.LevelLimits[i].Level) >= minimum
		}
		raw, err := json.Marshal(config)
		if err != nil {
			return err
		}
		return db.Create(&pageConfig.Entity{PageType: pageConfig.TrustSettings, Config: string(raw)}).Error
	}
	// Preserve explicit permissions and all unrelated stored settings on reruns.
	var root map[string]json.RawMessage
	if err := json.Unmarshal([]byte(row.Config), &root); err != nil {
		return err
	}
	var limits []map[string]json.RawMessage
	if err := json.Unmarshal(root["levelLimits"], &limits); err != nil {
		return err
	}
	for _, limit := range limits {
		if limit == nil {
			return errors.New("invalid trust limits")
		}
		if _, ok := limit["allowAttachmentUpload"]; ok {
			continue
		}
		var level int
		if err := json.Unmarshal(limit["level"], &level); err != nil {
			return err
		}
		limit["allowAttachmentUpload"], _ = json.Marshal(level >= minimum)
	}
	root["levelLimits"], _ = json.Marshal(limits)
	raw, err := json.Marshal(root)
	if err != nil {
		return err
	}
	return db.Model(&pageConfig.Entity{}).Where("id = ?", row.Id).Update("config", string(raw)).Error
}
