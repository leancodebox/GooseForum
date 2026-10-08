package api

import (
	"errors"
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"gorm.io/gorm"
)

func TestAgentSettingsPersistAndInvalidateCache(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&pageConfig.Entity{}); err != nil {
		t.Fatal(err)
	}
	old := pageConfig.GetByPageType(pageConfig.AgentSettings)
	t.Cleanup(func() {
		if old.Id == 0 {
			db.Where("page_type = ?", pageConfig.AgentSettings).Delete(&pageConfig.Entity{})
		} else {
			_ = pageConfig.SaveConfig(pageConfig.AgentSettings, old.Config)
		}
		hotdataserve.ClearAgentSettingsConfigCache()
	})
	_ = hotdataserve.GetAgentSettingsConfigCache()
	settings := pageConfig.DefaultAgentSettings()
	settings.Enabled = false
	settings.ManualTokens = false
	settings.ReadPerMinute = 17
	response := SaveAgentSettings(component.BetterRequest[SaveAgentSettingsReq]{Params: SaveAgentSettingsReq{Settings: settings}})
	if response.Data.Code != component.SUCCESS {
		t.Fatalf("save: %#v", response.Data)
	}
	if got := hotdataserve.GetAgentSettingsConfigCache(); got != settings {
		t.Fatalf("stale config: %#v", got)
	}
	settings.IPPerMinute = 0
	response = SaveAgentSettings(component.BetterRequest[SaveAgentSettingsReq]{Params: SaveAgentSettingsReq{Settings: settings}})
	if response.Data.Code != component.FAIL {
		t.Fatal("accepted invalid rate")
	}
	loaded, err := pageConfig.GetAgentSettings()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.IPPerMinute == 0 {
		t.Fatal("invalid rate persisted")
	}
	if err := pageConfig.SaveConfig(pageConfig.AgentSettings, `{"enabled":false,"manualTokens":false}`); err != nil {
		t.Fatal(err)
	}
	got, err := pageConfig.GetAgentSettings()
	if err != nil {
		t.Fatal(err)
	}
	if got.Enabled || got.ManualTokens || got.ReadPerMinute != 60 || got.IPPerMinute != 120 {
		t.Fatalf("missing-field defaults or explicit false lost: %#v", got)
	}
	if err := pageConfig.SaveConfig(pageConfig.AgentSettings, `{"enabled":"invalid"}`); err != nil {
		t.Fatal(err)
	}
	hotdataserve.ClearAgentSettingsConfigCache()
	if cfg := hotdataserve.GetAgentSettingsConfigCache(); cfg.Enabled || cfg.ManualTokens {
		t.Fatal("malformed configuration enabled Agent access")
	}
	if response := GetAgentSettings(component.BetterRequest[component.Null]{}); response.Data.Code != component.FAIL {
		t.Fatal("administrator cannot distinguish a configuration error from defaults")
	}
	if err := db.Callback().Query().Before("gorm:query").Register("agent_settings_read_failure", func(tx *gorm.DB) {
		if _, ok := tx.Statement.Model.(*pageConfig.Entity); ok {
			tx.AddError(errors.New("settings store unavailable"))
		}
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Callback().Query().Remove("agent_settings_read_failure") })
	hotdataserve.ClearAgentSettingsConfigCache()
	if cfg := hotdataserve.GetAgentSettingsConfigCache(); cfg.Enabled || cfg.ManualTokens {
		t.Fatal("database failure enabled Agent access")
	}
}
