package api

import (
	"fmt"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/accountrestrictions"
	"github.com/leancodebox/GooseForum/app/models/forum/oidcProviderStore"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/role"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/accountrestrictionservice"
	"github.com/leancodebox/GooseForum/app/service/adminpolicyservice"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"github.com/leancodebox/GooseForum/app/service/rolemanagementservice"
	"gorm.io/gorm"
)

func rolePolicyFixture(t *testing.T) (*gorm.DB, []users.EntityComplete) {
	t.Helper()
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&role.Entity{}, &rolePermissionRs.Entity{}, &users.EntityComplete{}, &accountrestrictions.History{}, &oidcProviderStore.TokenEntity{}, &oidcProviderStore.AuthorizationCodeEntity{}, &pageConfig.Entity{}); err != nil {
		t.Fatal(err)
	}
	var actors []users.EntityComplete
	for index, grant := range []permission.Enum{permission.Admin, permission.Admin, permission.RoleManager} {
		entity := role.Entity{RoleName: fmt.Sprintf("role-policy-%d-%d", time.Now().UnixNano(), index), Effective: 1}
		if err := db.Create(&entity).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Create(&rolePermissionRs.Entity{RoleId: entity.Id, PermissionId: grant.Id(), Effective: 1}).Error; err != nil {
			t.Fatal(err)
		}
		actor := users.EntityComplete{Username: entity.RoleName, Email: entity.RoleName + "@example.test", RoleId: entity.Id, TokenVersion: 1, IsActivated: 1}
		if err := db.Create(&actor).Error; err != nil {
			t.Fatal(err)
		}
		actors = append(actors, actor)
	}
	t.Cleanup(func() {
		for _, actor := range actors {
			db.Unscoped().Delete(&actor)
			db.Unscoped().Where("role_id = ?", actor.RoleId).Delete(&rolePermissionRs.Entity{})
			db.Unscoped().Delete(&role.Entity{}, actor.RoleId)
			db.Where("user_id = ?", actor.Id).Delete(&accountrestrictions.History{})
			permission.InvalidateRole(actor.RoleId)
		}
	})
	return db, actors
}

func TestRoleHandlersRejectSelfEscalationAndAdministratorChanges(t *testing.T) {
	db, actors := rolePolicyFixture(t)
	manager := actors[2]
	for _, request := range []RoleSaveReq{{Id: uint(manager.RoleId), RoleName: "self-admin", Permissions: []uint64{0}}, {RoleName: "new-admin", Permissions: []uint64{0}}, {Id: uint(actors[0].RoleId), RoleName: "remove-admin", Permissions: []uint64{permission.RoleManager.Id()}}} {
		response := RoleSave(component.BetterRequest[RoleSaveReq]{UserId: manager.Id, Params: request})
		if response.Data.MessageCode != component.MessageAdminRestrictionProtected {
			t.Fatalf("unsafe handler save accepted: %+v", response)
		}
	}
	response := RoleDel(component.BetterRequest[RoleSaveDel]{UserId: manager.Id, Params: RoleSaveDel{Id: uint(actors[0].RoleId)}})
	if response.Data.MessageCode != component.MessageAdminRestrictionProtected {
		t.Fatalf("unsafe handler delete accepted: %+v", response)
	}
	var grants []rolePermissionRs.Entity
	db.Where("role_id = ? AND effective = ?", manager.RoleId, 1).Find(&grants)
	if len(grants) != 1 || grants[0].PermissionId != permission.RoleManager.Id() {
		t.Fatal("rejected handler changed privileges")
	}
}

func TestRoleAndAccountChangesShareLastAdministratorProtection(t *testing.T) {
	_, actors := rolePolicyFixture(t)
	a, b := actors[0], actors[1]
	results := []error{
		accountrestrictionservice.Edit(0, accountrestrictionservice.Change{UserId: a.Id, Status: users.RestrictionBanned, Reason: "restriction", RoleId: a.RoleId, Validate: 1}),
		rolemanagementservice.Delete(a.Id, b.RoleId),
	}
	success := 0
	for _, err := range results {
		if err == nil {
			success++
		}
	}
	if success != 1 {
		t.Fatalf("concurrent role/account changes committed=%d; want one", success)
	}
	snapshot, err := adminpolicyservice.ReadSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	active := 0
	admins, err := users.AccountsInRoles(snapshot.AdminRoleIDs)
	if err != nil {
		t.Fatal(err)
	}
	for _, admin := range admins {
		if admin.EffectiveRestriction(time.Now()) == users.RestrictionNormal {
			active++
		}
	}
	if active != 1 {
		t.Fatalf("remaining normal administrators=%d; want one", active)
	}
}
