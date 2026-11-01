package pageConfig

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestTrustConfigValidation(t *testing.T) {
	c := DefaultTrustSettings()
	if c.LevelLimits[0].MaxDailyTopics == nil || *c.LevelLimits[0].MaxDailyTopics != 3 {
		t.Fatal("missing newcomer topic quota")
	}
	if !c.Valid() {
		t.Fatal("invalid defaults")
	}
	raw, _ := json.Marshal(c)
	if _, err := DecodeTrustSettings(raw); err != nil {
		t.Fatal(err)
	}
	c.PromotionRules[1].MinVisitDays = 100
	if !c.Valid() {
		t.Fatal("visit days unexpectedly capped at seven")
	}
	c.PromotionRules[1].MinTopicsVisited = 21
	if c.Valid() {
		t.Fatal("unbounded topic collection accepted")
	}
	c = DefaultTrustSettings()
	c.LevelLimits[0].MaxImagesPerPost = new(0)
	if !c.Valid() {
		t.Fatal("zero image limit rejected")
	}
	c.LevelLimits[0].MaxImagesPerPost = new(-1)
	if c.Valid() {
		t.Fatal("negative limit accepted")
	}
	if _, err := DecodeTrustSettings([]byte(`{"enabled":true}`)); err == nil {
		t.Fatal("unknown total switch accepted")
	}
	c = DefaultTrustSettings()
	c.PromotionRules[1] = c.PromotionRules[0]
	if c.Valid() {
		t.Fatal("duplicate target accepted")
	}
}

func TestRetiredPromotionSwitchIsIgnoredAndNotReturned(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		raw, _ := json.Marshal(DefaultTrustSettings())
		var fields map[string]any
		if err := json.Unmarshal(raw, &fields); err != nil {
			t.Fatal(err)
		}
		fields["autoPromotionEnabled"] = enabled
		raw, _ = json.Marshal(fields)
		config, err := DecodeTrustSettings(raw)
		if err != nil {
			t.Fatal(err)
		}
		if !config.PromotionRules[0].Enabled || !config.PromotionRules[1].Enabled {
			t.Fatal("retired switch changed target rules")
		}
		encoded, _ := json.Marshal(config)
		if bytes.Contains(encoded, []byte("autoPromotionEnabled")) {
			t.Fatal("retired switch returned")
		}
	}
}

func TestLegacyTrustSettingsDefaultProfileCover(t *testing.T) {
	raw, _ := json.Marshal(DefaultTrustSettings())
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		t.Fatal(err)
	}
	for _, item := range root["levelLimits"].([]any) {
		delete(item.(map[string]any), "allowProfileCover")
	}
	raw, _ = json.Marshal(root)
	config, err := DecodeTrustSettings(raw)
	if err != nil {
		t.Fatal(err)
	}
	for i, limit := range config.LevelLimits {
		if limit.AllowProfileCover != (i > 0) {
			t.Fatalf("level %d: unexpected permission", limit.Level)
		}
	}
	config.LevelLimits[0].AllowProfileCover = true
	config.LevelLimits[1].AllowProfileCover = false
	if issue := config.ValidateLimitOrder(); issue == nil || issue.Field != "allowProfileCover" {
		t.Fatalf("expected permission ordering error: %+v", issue)
	}
}

func TestTrustSettingsRejectNullLimit(t *testing.T) {
	raw, _ := json.Marshal(DefaultTrustSettings())
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		t.Fatal(err)
	}
	root["levelLimits"].([]any)[0] = nil
	raw, _ = json.Marshal(root)
	if _, err := DecodeTrustSettings(raw); err == nil {
		t.Fatal("null limit accepted")
	}
}
