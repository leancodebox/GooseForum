package datamigration

import (
	"encoding/json"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"gorm.io/gorm"
)

func TestParticipationMigrationPreservesEffectivePolicy(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(map[bool]string{false: "disabled", true: "enabled"}[enabled], func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			if err != nil {
				t.Fatal(err)
			}
			if err := db.AutoMigrate(&pageConfig.Entity{}); err != nil {
				t.Fatal(err)
			}
			posting := pageConfig.Entity{PageType: pageConfig.PostingSettings, Config: `{"textControl":{"minPostLength":9,"maxDailyTopicsPerUser":2,"newUserPostCooldownMinutes":30},"uploadControl":{"allowAttachments":false,"maxAttachmentSizeKb":300,"maxDailyUploadsPerUser":8,"newUserUploadCooldownMinutes":1440}}`}
			if err := db.Create(&posting).Error; err != nil {
				t.Fatal(err)
			}
			legacy := pageConfig.DefaultTrustSettings()
			legacy.SchemaVersion = 1
			raw, _ := json.Marshal(legacy)
			var values map[string]any
			_ = json.Unmarshal(raw, &values)
			values["restrictionsEnabled"] = enabled
			raw, _ = json.Marshal(values)
			trust := pageConfig.Entity{PageType: pageConfig.TrustSettings, Config: string(raw)}
			if err := db.Create(&trust).Error; err != nil {
				t.Fatal(err)
			}
			if err := MigrateParticipationSettingsWithDB(db); err != nil {
				t.Fatal(err)
			}
			if err := db.First(&trust, trust.Id).Error; err != nil {
				t.Fatal(err)
			}
			migrated, err := pageConfig.DecodeTrustSettings([]byte(trust.Config))
			if err != nil {
				t.Fatal(err)
			}
			for _, l := range migrated.LevelLimits {
				if l.MaxDailyTopics == nil || *l.MaxDailyTopics != 2 || l.MaxDailyUploads == nil || *l.MaxDailyUploads != 8 || l.MinPostAccountAgeMinutes != 30 || l.MinUploadAccountAgeMinutes != 1440 {
					t.Fatalf("lost quotas: %+v", l)
				}
				if !enabled && (!l.AllowImageUpload || l.MaxImagesPerPost != nil || l.MaxDailyReplies != nil) {
					t.Fatalf("disabled restrictions activated: %+v", l)
				}
			}
			if enabled && (migrated.LevelLimits[0].MaxImagesPerPost == nil || *migrated.LevelLimits[0].MaxImagesPerPost != 1) {
				t.Fatal("enabled limit lost")
			}
			first := trust.Config
			if err := MigrateParticipationSettingsWithDB(db); err != nil {
				t.Fatal(err)
			}
			db.First(&trust, trust.Id)
			if trust.Config != first {
				t.Fatal("rerun changed migrated policy")
			}
			db.First(&posting, posting.Id)
			var clean map[string]map[string]json.RawMessage
			if err := json.Unmarshal([]byte(posting.Config), &clean); err != nil {
				t.Fatal(err)
			}
			if _, ok := clean["textControl"]["maxDailyTopicsPerUser"]; ok {
				t.Fatal("duplicate quota retained")
			}
			if _, ok := clean["uploadControl"]["newUserUploadCooldownMinutes"]; ok {
				t.Fatal("duplicate age retained")
			}
			if string(clean["uploadControl"]["allowAttachments"]) != "false" || string(clean["textControl"]["minPostLength"]) != "9" {
				t.Fatal("unrelated basic settings changed")
			}
		})
	}
}
