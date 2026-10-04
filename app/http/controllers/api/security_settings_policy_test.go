package api

import (
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/accountrestrictionservice"
	"github.com/leancodebox/GooseForum/app/service/adminpolicyservice"
	"gorm.io/gorm"
)

func preserveSecuritySettings(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.AutoMigrate(&pageConfig.Entity{}); err != nil {
		t.Fatal(err)
	}
	previous := pageConfig.GetByPageType(pageConfig.SecuritySettings)
	t.Cleanup(func() {
		if previous.Id != 0 {
			if err := pageConfig.SaveConfig(pageConfig.SecuritySettings, previous.Config); err != nil {
				t.Error(err)
			}
		} else if err := db.Where("page_type = ?", pageConfig.SecuritySettings).Delete(&pageConfig.Entity{}).Error; err != nil {
			t.Error(err)
		}
		hotdataserve.ClearSecuritySettingsConfigCache()
	})
	if err := pageConfig.SaveConfig(pageConfig.SecuritySettings, `{"enableSignup":true,"enableEmailVerification":false}`); err != nil {
		t.Fatal(err)
	}
	hotdataserve.ClearSecuritySettingsConfigCache()
}

func TestEmailVerificationSettingsPreserveUsableAdministrator(t *testing.T) {
	db, actors := rolePolicyFixture(t)
	preserveSecuritySettings(t, db)
	for _, actor := range actors[:2] {
		if err := db.Model(&users.EntityComplete{}).Where("id = ?", actor.Id).Update("is_activated", users.ActivationPending).Error; err != nil {
			t.Fatal(err)
		}
	}
	request := component.BetterRequest[SaveSecuritySettingsReq]{UserId: actors[0].Id, Params: SaveSecuritySettingsReq{Settings: pageConfig.SecurityAndRegistration{EnableSignup: true, EnableEmailVerification: true}}}
	if response := SaveSecuritySettings(request); response.Data.MessageCode != component.MessageAdminRestrictionProtected {
		t.Fatalf("mandatory verification locked out all administrators: %+v", response)
	}
	if err := db.Model(&users.EntityComplete{}).Where("id = ?", actors[1].Id).Update("is_activated", users.ActivationSuccess).Error; err != nil {
		t.Fatal(err)
	}
	if response := SaveSecuritySettings(request); response.Data.Code != component.SUCCESS {
		t.Fatalf("verified administrator did not permit enabling verification: %+v", response)
	}
	if !hotdataserve.GetSecuritySettingsConfigCache().EnableEmailVerification {
		t.Fatal("verification setting not persisted")
	}
}

func TestSequentialVerificationPolicyAndRestrictionPreserveUsableAdministrator(t *testing.T) {
	db, actors := rolePolicyFixture(t)
	preserveSecuritySettings(t, db)
	if err := db.Model(&users.EntityComplete{}).Where("id = ?", actors[1].Id).Update("is_activated", users.ActivationPending).Error; err != nil {
		t.Fatal(err)
	}
	actor := actors[0]
	results := []error{
		adminpolicyservice.SaveSecuritySettings(pageConfig.SecurityAndRegistration{EnableSignup: true, EnableEmailVerification: true}),
		accountrestrictionservice.Edit(0, accountrestrictionservice.Change{UserId: actor.Id, RoleId: actor.RoleId, Validate: users.ActivationSuccess, Status: users.RestrictionBanned, Reason: "verification policy"}),
	}
	success := 0
	for _, err := range results {
		if err == nil {
			success++
		}
	}
	if success != 1 {
		t.Fatalf("verification enable/administrator ban committed=%d; want one", success)
	}
	snapshot, err := adminpolicyservice.ReadSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	remaining, err := users.HasNormalAdministrator(snapshot.AdminRoleIDs, 0, 0, time.Now(), snapshot.EmailVerificationEnabled)
	if err != nil || !remaining {
		t.Fatalf("concurrent operations left no usable administrator: remaining=%v err=%v", remaining, err)
	}
}
