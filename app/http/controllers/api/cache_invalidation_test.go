package api

import (
	"errors"
	"slices"
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/defaultconfig"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/role"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"github.com/leancodebox/GooseForum/app/service/themeservice"
	"gorm.io/gorm"
)

func TestRoleSaveRevokesWarmPermissionCache(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&role.Entity{}, &rolePermissionRs.Entity{}); err != nil {
		t.Fatal(err)
	}
	const roleID uint64 = 980901
	cleanup := func() {
		db.Unscoped().Where("role_id = ?", roleID).Delete(&rolePermissionRs.Entity{})
		db.Unscoped().Delete(&role.Entity{}, roleID)
		permission.InvalidateRole(roleID)
	}
	cleanup()
	t.Cleanup(cleanup)
	if err := db.Create(&role.Entity{Id: roleID, RoleName: "cache-revocation", Effective: 1}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&rolePermissionRs.Entity{RoleId: roleID, PermissionId: permission.Admin.Id(), Effective: 1}).Error; err != nil {
		t.Fatal(err)
	}
	if !permission.CheckRole(roleID, permission.Admin) {
		t.Fatal("administrator permission was not cached")
	}
	response := RoleSave(component.BetterRequest[RoleSaveReq]{Params: RoleSaveReq{
		Id: uint(roleID), RoleName: "cache-revocation", Permissions: []uint64{permission.SiteManager.Id()},
	}})
	if response.Data.Code != component.SUCCESS {
		t.Fatalf("role save failed: %+v", response)
	}
	if permission.CheckRole(roleID, permission.Admin) {
		t.Fatal("administrator permission remained after revocation")
	}
	if !permission.CheckRole(roleID, permission.SiteManager) {
		t.Fatal("replacement permission was not loaded")
	}
	if ids := rolePermissionRs.GetRsGroupByRoleIds([]uint64{roleID})[roleID]; slices.Contains(ids, permission.Admin.Id()) || !slices.Contains(ids, permission.SiteManager.Id()) {
		t.Fatalf("batch permission query retained revoked grant: %v", ids)
	}
	response = RoleDel(component.BetterRequest[RoleSaveDel]{Params: RoleSaveDel{Id: uint(roleID)}})
	if response.Data.Code != component.SUCCESS || permission.CheckAnyRole(roleID) {
		t.Fatalf("deleted role retained permissions: %+v", response)
	}
}

func TestThemeAndOIDCSettingsRejectFailedWrites(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&pageConfig.Entity{}); err != nil {
		t.Fatal(err)
	}
	theme := defaultconfig.GetDefaultSiteThemeConfig()
	theme.Enabled = false
	theme.Prepublish = &pageConfig.SiteThemePrepublish{Enabled: true, Themes: theme.Themes}
	cases := []struct {
		name     string
		pageType string
		config   any
		save     func() component.Response
	}{
		{"theme-save", pageConfig.SiteTheme, theme, func() component.Response {
			return SaveSiteTheme(component.BetterRequest[SaveSiteThemeReq]{Params: SaveSiteThemeReq{
				Settings: pageConfig.SiteThemeConfig{Enabled: true, Themes: theme.Themes},
			}})
		}},
		{"theme-publish", pageConfig.SiteTheme, theme, func() component.Response {
			return PublishSiteTheme(component.BetterRequest[component.Null]{})
		}},
		{"oidc-settings", pageConfig.OIDCProvider, pageConfig.OIDCProviderSettingsConfig{Enabled: false}, func() component.Response {
			return SaveOIDCProviderSettings(component.BetterRequest[SaveOIDCProviderSettingsReq]{Params: SaveOIDCProviderSettingsReq{Enabled: true}})
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			original := pageConfig.GetByPageType(test.pageType)
			t.Cleanup(func() {
				if original.Id != 0 {
					if err := pageConfig.SaveConfig(test.pageType, original.Config); err != nil {
						t.Error(err)
					}
				} else if err := db.Where("page_type = ?", test.pageType).Delete(&pageConfig.Entity{}).Error; err != nil {
					t.Error(err)
				}
				themeservice.ClearCaches()
			})
			stored := jsonopt.Encode(test.config)
			if err := pageConfig.SaveConfig(test.pageType, stored); err != nil {
				t.Fatal(err)
			}
			themeservice.ClearCaches()
			if test.pageType == pageConfig.SiteTheme {
				hotdataserve.GetSiteThemeConfigCache()
			}
			const hook = "test:reject-theme-and-oidc-update"
			if err := db.Callback().Update().Before("gorm:update").Register(hook, func(tx *gorm.DB) {
				if tx.Statement.Table == (&pageConfig.Entity{}).TableName() {
					tx.AddError(errors.New("injected settings update failure"))
				}
			}); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { db.Callback().Update().Remove(hook) })
			response := test.save()
			if response.Data.Code != component.FAIL {
				t.Fatalf("controller swallowed write failure: %+v", response)
			}
			if actual := pageConfig.GetByPageType(test.pageType); actual.Config != stored {
				t.Fatalf("failed write changed settings: %s", actual.Config)
			}
			if test.pageType == pageConfig.SiteTheme {
				if actual := hotdataserve.GetSiteThemeConfigCache(); jsonopt.Encode(actual) != stored {
					t.Fatalf("failed write changed warm theme cache: %+v", actual)
				}
			}
		})
	}
}

func TestPageConfigSaveRefreshesWarmCache(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&pageConfig.Entity{}); err != nil {
		t.Fatal(err)
	}
	original := pageConfig.GetByPageType(pageConfig.SiteSettings)
	t.Cleanup(func() {
		if original.Id != 0 {
			if err := pageConfig.SaveConfig(pageConfig.SiteSettings, original.Config); err != nil {
				t.Error(err)
			}
		} else {
			db.Unscoped().Where("page_type = ?", pageConfig.SiteSettings).Delete(&pageConfig.Entity{})
		}
		hotdataserve.ClearSiteSettingsConfigCache()
	})
	before := pageConfig.SiteSettingsConfig{SiteName: "before-cache-save"}
	if err := pageConfig.SaveConfig(pageConfig.SiteSettings, jsonopt.Encode(before)); err != nil {
		t.Fatal(err)
	}
	hotdataserve.ClearSiteSettingsConfigCache()
	if got := hotdataserve.GetSiteSettingsConfigCache(); got.SiteName != before.SiteName {
		t.Fatalf("warm config = %+v", got)
	}
	after := pageConfig.SiteSettingsConfig{SiteName: "after-cache-save"}
	response := savePageConfig(pageConfig.SiteSettings, after, hotdataserve.ClearSiteSettingsConfigCache)
	if response.Data.Code != component.SUCCESS {
		t.Fatalf("config save failed: %+v", response)
	}
	if got := hotdataserve.GetSiteSettingsConfigCache(); got.SiteName != after.SiteName {
		t.Fatalf("config remained stale after save: %+v", got)
	}
	const hook = "test:reject-page-config-update"
	if err := db.Callback().Update().Before("gorm:update").Register(hook, func(tx *gorm.DB) {
		if tx.Statement.Table == (&pageConfig.Entity{}).TableName() {
			tx.AddError(errors.New("injected config update failure"))
		}
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Callback().Update().Remove(hook) })
	cleared := false
	response = savePageConfig(pageConfig.SiteSettings, before, func() {
		cleared = true
		hotdataserve.ClearSiteSettingsConfigCache()
	})
	if response.Data.Code != component.FAIL || cleared {
		t.Fatalf("failed write reported success or invalidated cache: response=%+v cleared=%v", response, cleared)
	}
	if got := hotdataserve.GetSiteSettingsConfigCache(); got.SiteName != after.SiteName {
		t.Fatalf("failed write changed cached config: %+v", got)
	}
	if got := pageConfig.GetByPageType(pageConfig.SiteSettings); got.Config != jsonopt.Encode(after) {
		t.Fatalf("failed write changed persisted config: %+v", got)
	}
	response = SaveSiteSettings(component.BetterRequest[SaveSiteSettingsReq]{Params: SaveSiteSettingsReq{Settings: before}})
	if response.Data.Code != component.FAIL {
		t.Fatalf("site settings controller swallowed persistence failure: %+v", response)
	}
}
