package api

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/userservice"
)

func TestUserPrivacyPersistsFalseAndPreservesProfile(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}); err != nil {
		t.Fatal(err)
	}
	user := users.EntityComplete{Username: "privacy-save", Nickname: "Keep my name", Bio: "Keep my bio"}
	if err := users.Create(&user); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(&user) })
	for _, show := range []bool{false, true} {
		result := EditUserPrivacy(component.BetterRequest[EditUserPrivacyReq]{UserId: user.Id, Params: EditUserPrivacyReq{ShowActivity: &show, ShowTopics: &show, ShowFollowing: &show}})
		if result.Data.Code != component.SUCCESS {
			t.Fatalf("save: %+v", result)
		}
		saved, err := users.Get(user.Id)
		if err != nil {
			t.Fatal(err)
		}
		if saved.HideActivity != !show || saved.HideTopics != !show || saved.HideFollowing != !show || saved.Nickname != user.Nickname || saved.Bio != user.Bio {
			t.Fatalf("saved user: %+v", saved)
		}
		cached, ok := userservice.GetUserInfo(user.Id)
		if !ok || cached.HideActivity != !show || cached.HideTopics != !show || cached.HideFollowing != !show {
			t.Fatal("privacy cache is stale")
		}
	}
}
