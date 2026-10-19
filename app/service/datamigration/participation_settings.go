package datamigration

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"gorm.io/gorm"
)

func MigrateParticipationSettings() error {
	return MigrateParticipationSettingsWithDB(dbconnect.Connect())
}

func MigrateParticipationSettingsWithDB(db *gorm.DB) error {
	var posting, trust pageConfig.Entity
	if err := db.Where("page_type = ?", pageConfig.PostingSettings).First(&posting).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err := db.Where("page_type = ?", pageConfig.TrustSettings).First(&trust).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	var old struct {
		TextControl struct {
			Daily *int `json:"maxDailyTopicsPerUser"`
			Age   int  `json:"newUserPostCooldownMinutes"`
		} `json:"textControl"`
		UploadControl struct {
			Daily int `json:"maxDailyUploadsPerUser"`
			Age   int `json:"newUserUploadCooldownMinutes"`
		} `json:"uploadControl"`
	}
	old.TextControl.Daily = new(10)
	old.UploadControl.Daily = 10
	old.UploadControl.Age = 1440
	var root map[string]json.RawMessage
	if posting.Id > 0 {
		old.UploadControl.Daily = 0
		old.UploadControl.Age = 0
		if err := json.Unmarshal([]byte(posting.Config), &old); err != nil {
			return fmt.Errorf("legacy posting settings: %w", err)
		}
		if err := json.Unmarshal([]byte(posting.Config), &root); err != nil {
			return err
		}
	}
	legacy := struct {
		pageConfig.TrustConfig
		RestrictionsEnabled bool `json:"restrictionsEnabled"`
	}{TrustConfig: pageConfig.DefaultTrustSettings()}
	if trust.Id > 0 {
		if err := json.Unmarshal([]byte(trust.Config), &legacy); err != nil {
			return err
		}
	}
	if trust.Id == 0 || legacy.SchemaVersion == 1 {
		c := legacy.TrustConfig
		c.SchemaVersion = 3
		c.RegistrationCooldown = pageConfig.RegistrationCooldown{MinPostAccountAgeMinutes: max(0, old.TextControl.Age), MinUploadAccountAgeMinutes: max(0, old.UploadControl.Age)}
		for i := range c.LevelLimits {
			l := &c.LevelLimits[i]
			if !legacy.RestrictionsEnabled {
				l.AllowImageUpload = true
				l.MaxImagesPerPost = nil
				l.MaxExternalLinksPerPost = nil
				l.MaxMentionsPerPost = nil
				l.MaxDailyReplies = nil
				l.MaxDailyTopics = nil
			}
			if old.TextControl.Daily != nil && *old.TextControl.Daily > 0 && (l.MaxDailyTopics == nil || *old.TextControl.Daily < *l.MaxDailyTopics) {
				l.MaxDailyTopics = new(*old.TextControl.Daily)
			}
			l.MaxDailyUploads = nil
			if old.UploadControl.Daily > 0 {
				l.MaxDailyUploads = new(old.UploadControl.Daily)
			}
		}
		if !c.Valid() {
			return errors.New("invalid migrated trust settings")
		}
		raw, err := json.Marshal(c)
		if err != nil {
			return err
		}
		trust.PageType = pageConfig.TrustSettings
		trust.Config = string(raw)
		if err := db.Save(&trust).Error; err != nil {
			return err
		}
	} else if _, err := pageConfig.DecodeTrustSettings([]byte(trust.Config)); err != nil {
		return err
	}
	if posting.Id == 0 {
		return nil
	}
	for section, keys := range map[string][]string{"textControl": {"newUserPostCooldownMinutes", "maxDailyTopicsPerUser"}, "uploadControl": {"newUserUploadCooldownMinutes", "maxDailyUploadsPerUser"}} {
		raw, ok := root[section]
		if !ok {
			continue
		}
		var values map[string]json.RawMessage
		if err := json.Unmarshal(raw, &values); err != nil {
			return err
		}
		for _, k := range keys {
			delete(values, k)
		}
		encoded, err := json.Marshal(values)
		if err != nil {
			return err
		}
		root[section] = encoded
	}
	raw, err := json.Marshal(root)
	if err != nil {
		return err
	}
	posting.Config = string(raw)
	return db.Save(&posting).Error
}
