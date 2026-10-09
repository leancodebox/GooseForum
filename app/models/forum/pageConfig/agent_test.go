package pageConfig

import (
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"testing"
)

func TestAgentSettingsDefaultDisabledAndPreserveExplicitSettings(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&Entity{}); err != nil {
		t.Fatal(err)
	}
	old := GetByPageType(AgentSettings)
	t.Cleanup(func() {
		if old.Id == 0 {
			db.Where("page_type = ?", AgentSettings).Delete(&Entity{})
		} else {
			_ = SaveConfig(AgentSettings, old.Config)
		}
	})
	if err := db.Where("page_type = ?", AgentSettings).Delete(&Entity{}).Error; err != nil {
		t.Fatal(err)
	}
	config, err := GetAgentSettings()
	if err != nil || config.Enabled || config.ManualTokens || !config.Valid() {
		t.Fatalf("defaults: %+v %v", config, err)
	}
	if err := SaveConfig(AgentSettings, `{"enabled":true,"manualTokens":true}`); err != nil {
		t.Fatal(err)
	}
	config, err = GetAgentSettings()
	if err != nil || !config.Enabled || !config.ManualTokens || !config.Valid() {
		t.Fatalf("explicit settings: %+v %v", config, err)
	}
}
