package datamigration

import (
	"encoding/json"
	"errors"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/defaultconfig"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"gorm.io/gorm"
)

func MigrateAttachments() error { return MigrateAttachmentsWithDB(dbconnect.Connect()) }
func MigrateAttachmentsWithDB(db *gorm.DB) error {
	var row pageConfig.Entity
	if err := db.Where("page_type = ?", pageConfig.PostingSettings).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(row.Config), &fields); err != nil {
		return err
	}
	if fields == nil {
		return errors.New("invalid posting settings")
	}
	if _, ok := fields["attachments"]; ok {
		return nil
	}
	raw, err := json.Marshal(defaultconfig.GetDefaultPostingSettingsConfig().Attachments)
	if err != nil {
		return err
	}
	fields["attachments"] = raw
	raw, err = json.Marshal(fields)
	if err != nil {
		return err
	}
	return db.Model(&pageConfig.Entity{}).Where("id = ?", row.Id).Update("config", string(raw)).Error
}
