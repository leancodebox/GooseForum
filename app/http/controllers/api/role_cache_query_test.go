package api

import (
	"fmt"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/userservice"
	"gorm.io/gorm"
)

func TestRoleDeleteInvalidatesMemberCachesWithoutPerMemberSelects(t *testing.T) {
	db, actors := rolePolicyFixture(t)
	var members []users.EntityComplete
	for index := range 25 {
		members = append(members, users.EntityComplete{Username: fmt.Sprintf("role-member-%d-%d", time.Now().UnixNano(), index), RoleId: actors[2].RoleId})
	}
	if err := db.Create(&members).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for _, member := range members {
			db.Unscoped().Delete(&member)
			userservice.InvalidateUserCaches(member.Id)
		}
	})
	for _, member := range members {
		if info, ok := userservice.GetUserInfo(member.Id); !ok || info.RoleId != actors[2].RoleId {
			t.Fatalf("member cache not warmed: %+v", info)
		}
	}
	fullReads := 0
	const callback = "test:role-member-select-count"
	if err := db.Callback().Query().After("gorm:query").Register(callback, func(tx *gorm.DB) {
		if _, fullUser := tx.Statement.Dest.(*users.EntityComplete); fullUser {
			fullReads++
		}
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Callback().Query().Remove(callback) })
	response := RoleDel(component.BetterRequest[RoleSaveDel]{UserId: actors[0].Id, Params: RoleSaveDel{Id: uint(actors[2].RoleId)}})
	if response.Data.Code != component.SUCCESS {
		t.Fatalf("role deletion failed: %+v", response)
	}
	if fullReads != 0 {
		t.Fatalf("role deletion eagerly queried %d full member profiles", fullReads)
	}
	if err := db.Callback().Query().Remove(callback); err != nil {
		t.Fatal(err)
	}
	for _, member := range members {
		if info, ok := userservice.GetUserInfo(member.Id); !ok || info.RoleId != 0 || info.TokenVersion != member.TokenVersion+1 {
			t.Fatalf("member cache retained deleted role: %+v", info)
		}
	}
}
