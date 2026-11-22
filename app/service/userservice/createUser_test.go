package userservice

import (
	"strings"
	"testing"

	db "github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/role"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
)

func setupCreateUserTestDB(t *testing.T) {
	t.Helper()
	conn := db.Connect()
	if err := conn.AutoMigrate(&users.EntityComplete{}, &userStatistics.Entity{}, &role.Entity{}, &rolePermissionRs.Entity{}); err != nil {
		t.Fatalf("migrate user tables: %v", err)
	}
	conn.Where("1 = 1").Delete(&userStatistics.Entity{})
	conn.Where("1 = 1").Delete(&users.EntityComplete{})
}

func TestCreateUserStoresNormalizedLocale(t *testing.T) {
	setupCreateUserTestDB(t)

	user, err := CreateUser("lang-user", "password", "lang@example.com", false, "en-US")
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if user.Locale != "en" {
		t.Fatalf("Locale = %q, want en", user.Locale)
	}
}

func TestCreateUserKeepsLocaleEmptyWhenMissing(t *testing.T) {
	setupCreateUserTestDB(t)

	user, err := CreateUser("empty-locale", "password", "empty@example.com", false)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if user.Locale != "" {
		t.Fatalf("Locale = %q, want empty", user.Locale)
	}
}

func TestGenerateName(t *testing.T) {
	for range 4 {
		if name := GenerateGooseNickname(); name == "" {
			t.Fatal("expected generated nickname")
		}
	}
}

func TestDefaultAvatarAppliesAtRegistration(t *testing.T) {
	setupCreateUserTestDB(t)
	conn := db.Connect()
	if err := conn.AutoMigrate(&pageConfig.Entity{}); err != nil {
		t.Fatal(err)
	}
	original := pageConfig.GetByPageType(pageConfig.SiteSettings)
	t.Cleanup(func() {
		if original.Id == 0 {
			conn.Where("page_type = ?", pageConfig.SiteSettings).Delete(&pageConfig.Entity{})
		} else if err := pageConfig.SaveConfig(pageConfig.SiteSettings, original.Config); err != nil {
			t.Error(err)
		}
		hotdataserve.ClearSiteSettingsConfigCache()
	})
	register := func(name, settings string) *users.EntityComplete {
		t.Helper()
		if err := pageConfig.SaveConfig(pageConfig.SiteSettings, settings); err != nil {
			t.Fatal(err)
		}
		hotdataserve.ClearSiteSettingsConfigCache()
		user, err := CreateUser(name, "password", name+"@example.com", false)
		if err != nil {
			t.Fatal(err)
		}
		return user
	}
	legacy := register("legacy-avatar", "{}")
	if !strings.HasPrefix(legacy.AvatarUrl, "/static/pic/") {
		t.Fatalf("expected built-in avatar, got %q", legacy.AvatarUrl)
	}
	custom := register("custom-avatar", `{"defaultAvatarFile":"2026/site.png"}`)
	if custom.GetWebAvatarUrl() != "/file/img/2026/site.png" {
		t.Fatalf("custom avatar = %q", custom.GetWebAvatarUrl())
	}
	empty := register("empty-avatar", `{"disableDefaultAvatar":true,"defaultAvatarFile":"2026/site.png"}`)
	if empty.AvatarUrl != "" || empty.GetWebAvatarUrl() != "" {
		t.Fatal("disabled assignment returned an image")
	}
	if got := hotdataserve.GetSiteSettingsConfigCache().EnabledAvatarFiles(); len(got) != 0 {
		t.Fatalf("manual selection with automatic assignment off: %v", got)
	}
	selected := register("pool-avatar", `{"avatarPool":[{"file":"2026/skipped.png","enabled":false},{"file":"/static/pic/3.webp","enabled":true}]}`)
	if selected.AvatarUrl != "/static/pic/3.webp" {
		t.Fatalf("selected avatar = %q", selected.AvatarUrl)
	}
	stored, err := users.Get(custom.Id)
	if err != nil || stored.AvatarUrl != custom.AvatarUrl {
		t.Fatalf("existing avatar changed: %q, %v", stored.AvatarUrl, err)
	}
}
