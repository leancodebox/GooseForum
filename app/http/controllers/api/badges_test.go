package api

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/badgeStats"
	"github.com/leancodebox/GooseForum/app/models/forum/badges"
	"github.com/leancodebox/GooseForum/app/models/forum/userBadges"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/badgeservice"
)

func TestBadgeRecipientPageUsesStoredTotalAndOmitsSearchTotal(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&badges.Entity{}, &badgeStats.Entity{}, &userBadges.Entity{}, &users.EntityComplete{}); err != nil {
		t.Fatal(err)
	}
	code := fmt.Sprintf("page_%d", time.Now().UnixNano())
	badge := badges.Entity{Code: code, Type: badges.TypeCustom, GrantMode: badges.GrantModeManual, Name: "Award", IsEnabled: true}
	if err := badges.Save(&badge); err != nil {
		t.Fatal(err)
	}
	badgeservice.InvalidateDefinitions()
	var ids []uint64
	t.Cleanup(func() {
		db.Where("badge_code = ?", code).Delete(&userBadges.Entity{})
		db.Where("badge_code = ?", code).Delete(&badgeStats.Entity{})
		db.Where("code = ?", code).Delete(&badges.Entity{})
		if len(ids) > 0 {
			db.Unscoped().Where("id IN ?", ids).Delete(&users.EntityComplete{})
		}
		badgeservice.InvalidateDefinitions()
	})
	for _, suffix := range []string{"alice", "bob"} {
		user := users.EntityComplete{Username: code + suffix}
		if err := db.Create(&user).Error; err != nil {
			t.Fatal(err)
		}
		ids = append(ids, user.Id)
		if _, err := userBadges.GrantManual(user.Id, code, userBadges.SourceManual, "thanks", 0, nil); err != nil {
			t.Fatal(err)
		}
	}
	for _, test := range []struct {
		search  string
		page    int
		hasNext bool
		total   bool
	}{
		{"", 1, true, true}, {"", 2, false, true}, {"alice", 1, false, false}, {"absent", 1, false, false},
	} {
		response := BadgeRecipients(component.BetterRequest[BadgeRecipientsReq]{Params: BadgeRecipientsReq{Code: code, Page: test.page, PageSize: 1, Search: test.search}})
		if response.Data.Code != component.SUCCESS {
			t.Fatalf("response=%+v", response)
		}
		result, ok := response.Data.Result.(BadgeRecipientsPage)
		if !ok {
			t.Fatalf("wrong page type: %T", response.Data.Result)
		}
		if result.HasNext != test.hasNext || (result.Total != nil) != test.total {
			t.Fatalf("page=%+v test=%+v", result, test)
		}
		if result.Total != nil && *result.Total != 2 {
			t.Fatalf("total=%d", *result.Total)
		}
		encoded, err := json.Marshal(result)
		if err != nil {
			t.Fatal(err)
		}
		var fields map[string]any
		if err := json.Unmarshal(encoded, &fields); err != nil {
			t.Fatal(err)
		}
		if _, present := fields["hasNext"]; !present {
			t.Fatal("false hasNext omitted")
		}
		if _, present := fields["total"]; present != test.total {
			t.Fatalf("search total must be omitted: %s", encoded)
		}
	}
}
