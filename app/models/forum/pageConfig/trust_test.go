package pageConfig

import (
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
