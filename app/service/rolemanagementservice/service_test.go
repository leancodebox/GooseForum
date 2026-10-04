package rolemanagementservice

import (
	"errors"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/role"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"gorm.io/gorm"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := dbconnect.Connect()
	if err := db.Migrator().DropTable(&users.EntityComplete{}, &role.Entity{}, &rolePermissionRs.Entity{}, &pageConfig.Entity{}); err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&users.EntityComplete{}, &role.Entity{}, &rolePermissionRs.Entity{}, &pageConfig.Entity{}); err != nil {
		t.Fatal(err)
	}
	for _, r := range []role.Entity{{Id: 1, RoleName: "admin-a", Effective: 1}, {Id: 2, RoleName: "admin-b", Effective: 1}, {Id: 3, RoleName: "role-manager", Effective: 1}} {
		if err := db.Create(&r).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, g := range []rolePermissionRs.Entity{{RoleId: 1, PermissionId: permission.Admin.Id(), Effective: 1}, {RoleId: 2, PermissionId: permission.Admin.Id(), Effective: 1}, {RoleId: 3, PermissionId: permission.RoleManager.Id(), Effective: 1}} {
		if err := db.Create(&g).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, u := range []users.EntityComplete{{Id: 1, Username: "role-admin-a", RoleId: 1, TokenVersion: 4}, {Id: 2, Username: "role-admin-b", RoleId: 2, TokenVersion: 4}, {Id: 3, Username: "role-manager", RoleId: 3, TokenVersion: 4}} {
		if err := db.Create(&u).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func TestRoleManagerCannotEscalateOrModifyAdministratorRole(t *testing.T) {
	testDB(t)
	for _, change := range []Change{{Id: 3, Name: "escalate", Permissions: []uint64{0}}, {Name: "new-admin", Permissions: []uint64{0}}, {Id: 1, Name: "downgrade", Permissions: []uint64{permission.RoleManager.Id()}}, {Id: 1, Delete: true}} {
		if _, _, err := apply(3, change, time.Now()); !errors.Is(err, ErrProtected) {
			t.Fatalf("unsafe role change accepted: %+v %v", change, err)
		}
	}
	if _, _, err := apply(3, Change{Id: 3, Name: "renamed manager", Permissions: []uint64{permission.RoleManager.Id()}}, time.Now()); err != nil {
		t.Fatal(err)
	}
}

func TestLastAdministratorRoleAndRollback(t *testing.T) {
	db := testDB(t)
	if _, _, err := apply(1, Change{Id: 2, Delete: true}, time.Now()); err != nil {
		t.Fatal(err)
	}
	state, _ := users.GetAccountState(2)
	if state.RoleId != 0 || state.TokenVersion != 5 {
		t.Fatal("deleted role retained user permission/session")
	}
	for _, change := range []Change{{Id: 1, Delete: true}, {Id: 1, Name: "last", Permissions: []uint64{permission.RoleManager.Id()}}} {
		if _, _, err := apply(1, change, time.Now()); !errors.Is(err, ErrProtected) {
			t.Fatalf("last administrator removed: %v", err)
		}
	}
	db.Callback().Create().Before("gorm:create").Register("test:role-grant-failure", func(tx *gorm.DB) {
		if tx.Statement.Table == "role_permission_rs" {
			tx.AddError(errors.New("grant failure"))
		}
	})
	if _, _, err := apply(1, Change{Id: 3, Name: "changed", Permissions: []uint64{permission.SiteManager.Id()}}, time.Now()); err == nil {
		t.Fatal("failed grant accepted")
	}
	var entity role.Entity
	db.First(&entity, 3)
	if entity.RoleName != "changed" {
		t.Fatal("successful name update was not retained")
	}
	var grants []rolePermissionRs.Entity
	db.Where("role_id = ?", 3).Find(&grants)
	if len(grants) != 0 {
		t.Fatal("failed grant creation left unexpected permissions")
	}
	db.Callback().Create().Remove("test:role-grant-failure")
	if _, _, err := apply(1, Change{Id: 3, Name: "changed", Permissions: []uint64{permission.SiteManager.Id()}}, time.Now()); err != nil {
		t.Fatalf("retry failed: %v", err)
	}
}

func TestSequentialAdministratorRoleDeletesPreserveOneAdministrator(t *testing.T) {
	db := testDB(t)
	var results []error
	for _, id := range []uint64{1, 2} {
		_, _, err := apply(id, Change{Id: id, Delete: true}, time.Now())
		results = append(results, err)
	}
	success := 0
	for _, err := range results {
		if err == nil {
			success++
		}
	}
	if success != 1 {
		t.Fatalf("role deletions=%d; want one", success)
	}
	var count int64
	db.Model(&users.EntityComplete{}).Where("role_id IN (SELECT role_id FROM role_permission_rs WHERE permission_id = ? AND effective = ? AND deleted_at IS NULL)", permission.Admin.Id(), 1).Count(&count)
	if count != 1 {
		t.Fatalf("remaining administrators=%d", count)
	}
}

func TestPendingVerificationDoesNotProtectAdministratorRoleRemoval(t *testing.T) {
	db := testDB(t)
	if err := db.Model(&users.EntityComplete{}).Where("id = ?", 2).Update("requires_email_verification", true).Error; err != nil {
		t.Fatal(err)
	}
	for _, change := range []Change{{Id: 1, Delete: true}, {Id: 1, Name: "downgraded", Permissions: []uint64{permission.RoleManager.Id()}}} {
		if _, _, err := apply(1, change, time.Now()); !errors.Is(err, ErrProtected) {
			t.Fatalf("only pending administrator remained: %+v err=%v", change, err)
		}
	}
}
