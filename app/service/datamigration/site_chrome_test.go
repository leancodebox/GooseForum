package datamigration

import (
	"testing"

	db "github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"gorm.io/gorm"
)

func TestMigrateSiteChromeContentCopiesBrandAndFooter(t *testing.T) {
	resetPageConfigDB(t)

	pageConfig.CreateOrSave(&pageConfig.Entity{
		PageType: pageConfig.SiteSettings,
		Config: jsonopt.Encode(legacySiteSettingsConfig{
			FooterInfo: pageConfig.FooterInfo{
				Primary: []pageConfig.PItem{{Content: "GooseForum © 2024"}},
				List:    []pageConfig.FooterItem{{Name: "RSS", Url: "/rss.xml"}},
			},
			BrandType:  "text",
			BrandText:  "Goose",
			BrandImage: "/brand.png",
		}),
	})
	pageConfig.CreateOrSave(&pageConfig.Entity{
		PageType: pageConfig.SiteChrome,
		Config: jsonopt.Encode(pageConfig.SiteChromeConfig{
			Header: []pageConfig.ChromeItem{{ID: "custom", Enabled: true, Type: "link", Label: "Custom", URL: "/custom"}},
		}),
	})

	result := MigrateSiteChromeContent()
	if result.Failed != 0 || !result.Migrated {
		t.Fatalf("MigrateSiteChromeContent() = %+v, want migrated with no failures", result)
	}

	got := pageConfig.GetConfigByPageType(pageConfig.SiteChrome, pageConfig.SiteChromeConfig{})
	if got.BrandType != "text" || got.BrandText != "Goose" || got.BrandImage != "/brand.png" {
		t.Fatalf("brand = %q %q %q, want migrated brand", got.BrandType, got.BrandText, got.BrandImage)
	}
	if len(got.FooterInfo.List) != 1 || got.FooterInfo.List[0].Name != "RSS" || got.FooterInfo.List[0].Url != "/rss.xml" {
		t.Fatalf("footer links = %#v, want migrated RSS link", got.FooterInfo.List)
	}
	if len(got.FooterInfo.Primary) != 1 || got.FooterInfo.Primary[0].Content != "GooseForum © 2024" {
		t.Fatalf("footer primary = %#v, want migrated primary text", got.FooterInfo.Primary)
	}
	if !hasChromeItem(got.Header, "custom") {
		t.Fatalf("header = %#v, want existing custom header preserved", got.Header)
	}
}

func TestMigrateSiteChromeContentKeepsDefaultsWithoutLegacySiteSettings(t *testing.T) {
	resetPageConfigDB(t)

	result := MigrateSiteChromeContent()
	if result.Failed != 0 || !result.Migrated {
		t.Fatalf("MigrateSiteChromeContent() = %+v, want migrated with no failures", result)
	}

	got := pageConfig.GetConfigByPageType(pageConfig.SiteChrome, pageConfig.SiteChromeConfig{})
	if len(got.FooterInfo.List) == 0 || len(got.FooterInfo.Primary) == 0 {
		t.Fatalf("footer = %#v, want default footer preserved", got.FooterInfo)
	}
	if got.BrandType != "default" {
		t.Fatalf("brand type = %q, want default", got.BrandType)
	}
	if !hasChromeItem(got.Header, "sponsors") || !hasChromeItem(got.Header, "links") {
		t.Fatalf("header = %#v, want default system links", got.Header)
	}
}

func resetPageConfigDB(t *testing.T) *gorm.DB {
	t.Helper()
	preferences.Set("db.default.connection", "sqlite")
	preferences.Set("db.default.path", ":memory:")
	conn := db.Connect()
	if err := conn.AutoMigrate(&pageConfig.Entity{}); err != nil {
		t.Fatalf("migrate page config: %v", err)
	}
	if err := conn.Exec("DELETE FROM page_config").Error; err != nil {
		t.Fatalf("clear page config: %v", err)
	}
	return conn
}

func hasChromeItem(items []pageConfig.ChromeItem, id string) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}
