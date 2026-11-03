package migration

import (
	"encoding/json"
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
)

func TestVersion28StartupMigratesAttachmentPermissions(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&pageConfig.Entity{}); err != nil {
		t.Fatal(err)
	}
	types := []string{pageConfig.Migration, pageConfig.TrustSettings, pageConfig.PostingSettings}
	var previous []pageConfig.Entity
	if err := db.Where("page_type IN ?", types).Find(&previous).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("page_type IN ?", types).Delete(&pageConfig.Entity{})
		if len(previous) > 0 {
			db.Create(&previous)
		}
	})
	raw, err := json.Marshal(pageConfig.DefaultTrustSettings())
	if err != nil {
		t.Fatal(err)
	}
	var legacy map[string]any
	if err := json.Unmarshal(raw, &legacy); err != nil {
		t.Fatal(err)
	}
	for _, limit := range legacy["levelLimits"].([]any) {
		delete(limit.(map[string]any), "allowAttachmentUpload")
	}
	raw, err = json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if err := pageConfig.SaveConfig(pageConfig.TrustSettings, string(raw)); err != nil {
		t.Fatal(err)
	}
	if err := pageConfig.SaveConfig(pageConfig.PostingSettings, `{"attachments":{"minTrustLevel":1000}}`); err != nil {
		t.Fatal(err)
	}
	if err := pageConfig.SyncMigrationVersion(28); err != nil {
		t.Fatal(err)
	}
	if err := runVersionedDataMigrations(); err != nil {
		t.Fatal(err)
	}
	if version := pageConfig.GetMigrationVersion(); version < 29 || version != pageConfig.AppMigrationVersion {
		t.Fatalf("startup skipped attachment migration: version=%d target=%d", version, pageConfig.AppMigrationVersion)
	}
	config, err := pageConfig.GetTrustSettings()
	if err != nil {
		t.Fatal(err)
	}
	for _, limit := range config.LevelLimits {
		if limit.AllowAttachmentUpload != (limit.Level >= 1000) {
			t.Fatalf("legacy threshold lost: %+v", limit)
		}
	}
}
