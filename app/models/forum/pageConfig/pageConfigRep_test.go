package pageConfig

import (
	"fmt"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
)

func TestCompareAndSwapConfigRejectsStaleValue(t *testing.T) {
	if err := dbconnect.Connect().AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("migrate page config: %v", err)
	}
	pageType := fmt.Sprintf("test-cas-%d", time.Now().UnixNano())
	if err := SaveConfig(pageType, `{"version":1}`); err != nil {
		t.Fatalf("save initial config: %v", err)
	}
	t.Cleanup(func() { builder().Where("page_type = ?", pageType).Delete(&Entity{}) })

	initial := GetByPageType(pageType)
	saved, err := CompareAndSwapConfig(initial, `{"version":2}`)
	if err != nil || !saved {
		t.Fatalf("first compare-and-swap = %v, %v", saved, err)
	}
	saved, err = CompareAndSwapConfig(initial, `{"version":3}`)
	if err != nil {
		t.Fatalf("stale compare-and-swap: %v", err)
	}
	if saved {
		t.Fatal("stale compare-and-swap overwrote current config")
	}
	if got := GetByPageType(pageType).Config; got != `{"version":2}` {
		t.Fatalf("config = %s", got)
	}
}

func TestLegacySiteSavePreservesPoolAndExplicitEmptyClearsIt(t *testing.T) {
	if err := dbconnect.Connect().AutoMigrate(&Entity{}); err != nil {
		t.Fatal(err)
	}
	old := GetByPageType(SiteSettings)
	t.Cleanup(func() {
		if old.Id == 0 {
			builder().Where("page_type = ?", SiteSettings).Delete(&Entity{})
		} else if err := SaveConfig(SiteSettings, old.Config); err != nil {
			t.Error(err)
		}
	})
	initial := SiteSettingsConfig{DefaultLocale: "it", DisableDefaultAvatar: true, AvatarPool: []AvatarOption{{File: "example.png", Enabled: true}}}
	if err := SaveSiteConfig(initial); err != nil {
		t.Fatal(err)
	}
	if err := SaveSiteConfig(SiteSettingsConfig{SiteName: "updated"}); err != nil {
		t.Fatal(err)
	}
	got := GetConfigByPageType(SiteSettings, SiteSettingsConfig{})
	if got.DefaultLocale != "it" || !got.DisableDefaultAvatar || len(got.AvatarPool) != 1 || got.AvatarPool[0] != initial.AvatarPool[0] || got.SiteName != "updated" {
		t.Fatalf("legacy save: %+v", got)
	}
	got.AvatarPool = []AvatarOption{}
	if err := SaveSiteConfig(got); err != nil {
		t.Fatal(err)
	}
	if pool := GetConfigByPageType(SiteSettings, SiteSettingsConfig{}).AvatarPool; pool == nil || len(pool) != 0 {
		t.Fatalf("explicit empty pool = %#v", pool)
	}
}
