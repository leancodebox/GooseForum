package datamigration

import (
	"encoding/json"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"gorm.io/gorm"
)

func TestSharedRegistrationCooldownMigration(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&pageConfig.Entity{}); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(pageConfig.DefaultTrustSettings())
	var old map[string]any
	if err := json.Unmarshal(raw, &old); err != nil {
		t.Fatal(err)
	}
	old["schemaVersion"] = 2
	delete(old, "registrationCooldown")
	for i, value := range old["levelLimits"].([]any) {
		limit := value.(map[string]any)
		limit["minPostAccountAgeMinutes"] = i * 10
		limit["minUploadAccountAgeMinutes"] = (3 - i) * 60
	}
	raw, _ = json.Marshal(old)
	row := pageConfig.Entity{PageType: pageConfig.TrustSettings, Config: string(raw)}
	if err := db.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
	if err := MigrateRegistrationCooldownWithDB(db); err != nil {
		t.Fatal(err)
	}
	if err := db.First(&row, row.Id).Error; err != nil {
		t.Fatal(err)
	}
	config, err := pageConfig.DecodeTrustSettings([]byte(row.Config))
	if err != nil {
		t.Fatal(err)
	}
	if config.SchemaVersion != 3 || config.RegistrationCooldown.MinPostAccountAgeMinutes != 30 || config.RegistrationCooldown.MinUploadAccountAgeMinutes != 180 {
		t.Fatalf("shared cooldown=%+v", config)
	}
	if config.LevelLimits[0].MaxDailyTopics == nil || *config.LevelLimits[0].MaxDailyTopics != 3 {
		t.Fatal("unrelated limit changed")
	}
	var backup pageConfig.Entity
	if err := db.Where("page_type = ?", "trustSettingsBeforeSharedCooldown").First(&backup).Error; err != nil {
		t.Fatal(err)
	}
	if backup.Config != string(raw) {
		t.Fatal("original policy not preserved")
	}
	// Later administrator changes must survive another migration attempt.
	config.RegistrationCooldown.MinPostAccountAgeMinutes = 5
	updated, _ := json.Marshal(config)
	if err := db.Model(&row).Update("config", string(updated)).Error; err != nil {
		t.Fatal(err)
	}
	if err := MigrateRegistrationCooldownWithDB(db); err != nil {
		t.Fatal(err)
	}
	if err := db.First(&row, row.Id).Error; err != nil {
		t.Fatal(err)
	}
	if row.Config != string(updated) {
		t.Fatal("rerun overwrote administrator policy")
	}
}
