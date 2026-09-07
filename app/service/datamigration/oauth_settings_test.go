package datamigration

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"gorm.io/gorm"
)

func TestMigrateLegacyOAuthSettings(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&pageConfig.Entity{}); err != nil {
		t.Fatal(err)
	}

	result := MigrateLegacyOAuthSettingsWithDB(db, " legacy-id ", " legacy-secret ")
	if result.Failed != 0 || !result.Migrated || result.Skipped {
		t.Fatalf("migration result = %#v", result)
	}

	var entity pageConfig.Entity
	if err := db.Where("page_type = ?", pageConfig.OAuthSettings).First(&entity).Error; err != nil {
		t.Fatal(err)
	}
	config := jsonopt.Decode[pageConfig.OAuthSettingsConfig](entity.Config)
	if !config.GitHub.Enabled || config.GitHub.ClientID != "legacy-id" || config.GitHub.ClientSecret != "legacy-secret" {
		t.Fatalf("migrated config = %#v", config)
	}

	result = MigrateLegacyOAuthSettingsWithDB(db, "replacement-id", "replacement-secret")
	if result.Failed != 0 || result.Migrated || !result.Skipped {
		t.Fatalf("second migration result = %#v", result)
	}
	if err := db.Where("page_type = ?", pageConfig.OAuthSettings).First(&entity).Error; err != nil {
		t.Fatal(err)
	}
	config = jsonopt.Decode[pageConfig.OAuthSettingsConfig](entity.Config)
	if config.GitHub.ClientID != "legacy-id" || config.GitHub.ClientSecret != "legacy-secret" {
		t.Fatalf("existing config was overwritten = %#v", config)
	}
}

func TestMigrateLegacyOAuthSettingsSkipsIncompleteCredentials(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&pageConfig.Entity{}); err != nil {
		t.Fatal(err)
	}

	result := MigrateLegacyOAuthSettingsWithDB(db, "legacy-id", "")
	if result.Failed != 0 || result.Migrated || !result.Skipped {
		t.Fatalf("migration result = %#v", result)
	}
	var count int64
	if err := db.Model(&pageConfig.Entity{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("created %d settings rows", count)
	}
}
