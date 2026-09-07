package datamigration

import (
	"errors"
	"strings"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/models/defaultconfig"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"gorm.io/gorm"
)

type OAuthSettingsMigrationResult struct {
	Migrated   bool
	Skipped    bool
	Failed     int
	LastFailed string
}

// MigrateLegacyOAuthSettings copies the former TOML GitHub credentials into
// the database-backed OAuth settings without overwriting an existing config.
func MigrateLegacyOAuthSettings() OAuthSettingsMigrationResult {
	result := MigrateLegacyOAuthSettingsWithDB(
		dbconnect.Connect(),
		preferences.GetString("github.client_id", ""),
		preferences.GetString("github.client_secret", ""),
	)
	if result.Migrated {
		hotdataserve.ClearOAuthSettingsConfigCache()
	}
	return result
}

func MigrateLegacyOAuthSettingsWithDB(db *gorm.DB, clientID, clientSecret string) OAuthSettingsMigrationResult {
	result := OAuthSettingsMigrationResult{}
	var existing pageConfig.Entity
	query := db.Where("page_type = ?", pageConfig.OAuthSettings).First(&existing)
	if query.Error != nil && !errors.Is(query.Error, gorm.ErrRecordNotFound) {
		result.Failed = 1
		result.LastFailed = query.Error.Error()
		return result
	}
	if existing.Id != 0 {
		result.Skipped = true
		return result
	}

	clientID = strings.TrimSpace(clientID)
	clientSecret = strings.TrimSpace(clientSecret)
	if clientID == "" || clientSecret == "" {
		result.Skipped = true
		return result
	}

	config := defaultconfig.GetDefaultOAuthSettingsConfig()
	config.GitHub = pageConfig.OAuthProviderConfig{
		Enabled:      true,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
	entity := pageConfig.Entity{PageType: pageConfig.OAuthSettings, Config: jsonopt.Encode(config)}
	if err := db.Create(&entity).Error; err != nil {
		result.Failed = 1
		result.LastFailed = err.Error()
		return result
	}
	result.Migrated = true
	return result
}
