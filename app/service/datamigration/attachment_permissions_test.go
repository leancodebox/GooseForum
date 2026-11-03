package datamigration

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"gorm.io/gorm"
)

func TestAttachmentPermissionMigration(t *testing.T) {
	for _, minimum := range []int{0, 100, 200, 1000} {
		for _, state := range []string{"absent", "legacy", "explicit"} {
			t.Run(fmt.Sprintf("%d/%s", minimum, state), func(t *testing.T) {
				db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
				if err != nil {
					t.Fatal(err)
				}
				if err := db.AutoMigrate(&pageConfig.Entity{}); err != nil {
					t.Fatal(err)
				}
				posting := pageConfig.Entity{PageType: pageConfig.PostingSettings, Config: fmt.Sprintf(`{"attachments":{"minTrustLevel":%d}}`, minimum)}
				if err := db.Create(&posting).Error; err != nil {
					t.Fatal(err)
				}
				if state != "absent" {
					raw, _ := json.Marshal(pageConfig.DefaultTrustSettings())
					var root map[string]any
					if err := json.Unmarshal(raw, &root); err != nil {
						t.Fatal(err)
					}
					if state == "legacy" {
						for _, limit := range root["levelLimits"].([]any) {
							delete(limit.(map[string]any), "allowAttachmentUpload")
						}
					}
					raw, _ = json.Marshal(root)
					if err := db.Create(&pageConfig.Entity{PageType: pageConfig.TrustSettings, Config: string(raw)}).Error; err != nil {
						t.Fatal(err)
					}
				}
				for range 2 {
					if err := MigrateAttachmentPermissionsWithDB(db); err != nil {
						t.Fatal(err)
					}
					var row pageConfig.Entity
					if err := db.Where("page_type = ?", pageConfig.TrustSettings).First(&row).Error; err != nil {
						t.Fatal(err)
					}
					config, err := pageConfig.DecodeTrustSettings([]byte(row.Config))
					if err != nil {
						t.Fatal(err)
					}
					for _, limit := range config.LevelLimits {
						threshold := minimum
						if state == "explicit" {
							threshold = 100
						}
						if limit.AllowAttachmentUpload != (int(limit.Level) >= threshold) {
							t.Fatalf("wrong permission: %+v", limit)
						}
					}
				}
			})
		}
	}
}
