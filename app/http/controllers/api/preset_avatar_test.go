package api

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
)

func TestManualAvatarFollowsAssignmentSwitchAndPool(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}, &pageConfig.Entity{}); err != nil {
		t.Fatal(err)
	}
	previous := pageConfig.GetByPageType(pageConfig.SiteSettings)
	t.Cleanup(func() {
		if previous.Id == 0 {
			db.Where("page_type = ?", pageConfig.SiteSettings).Delete(&pageConfig.Entity{})
		} else if err := pageConfig.SaveConfig(pageConfig.SiteSettings, previous.Config); err != nil {
			t.Error(err)
		}
		hotdataserve.ClearSiteSettingsConfigCache()
	})
	user := users.MakeUser("manual-preset", "password", "preset@example.test")
	if err := users.Create(user); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(user) })
	const file = "2026/preset.webp"
	config := pageConfig.SiteSettingsConfig{AvatarPool: []pageConfig.AvatarOption{{File: file, Enabled: true}}}
	savePool := func() {
		t.Helper()
		if err := pageConfig.SaveSiteConfig(config); err != nil {
			t.Fatal(err)
		}
		hotdataserve.ClearSiteSettingsConfigCache()
	}
	choose := func(file string) bool {
		return SetPresetAvatar(component.BetterRequest[SetPresetAvatarReq]{UserId: user.Id, Params: SetPresetAvatarReq{AvatarUrl: file}}).Data.Code == component.SUCCESS
	}
	savePool()
	if choose("/static/pic/1.webp") || !choose(file) {
		t.Fatal("manual selection ignored configured candidates")
	}
	config.DisableDefaultAvatar = true
	savePool()
	if choose(file) {
		t.Fatal("accepted a preset after automatic assignment was switched off")
	}
	config.DisableDefaultAvatar = false
	config.AvatarPool[0].Enabled = false
	savePool()
	if choose(file) {
		t.Fatal("accepted a disabled preset")
	}
	current, err := users.Get(user.Id)
	if err != nil || current.AvatarUrl != file {
		t.Fatalf("existing avatar = %q, %v", current.AvatarUrl, err)
	}
}
