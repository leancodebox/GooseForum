package trustlevelservice

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/trustlevel"
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/trusthistory"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
)

func TestReviewManualLockAndContentLimits(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}, &pageConfig.Entity{}, &trusthistory.History{}, &eventNotification.Entity{}, &topics.Entity{}, &posts.Entity{}); err != nil {
		t.Fatal(err)
	}
	config := pageConfig.DefaultTrustSettings()
	raw, _ := json.Marshal(config)
	previous := pageConfig.GetByPageType(pageConfig.TrustSettings)
	if err := pageConfig.SaveConfig(pageConfig.TrustSettings, string(raw)); err != nil {
		t.Fatal(err)
	}
	hotdataserve.ClearTrustSettingsCache()
	t.Cleanup(func() {
		if previous.Id > 0 {
			_ = pageConfig.SaveConfig(pageConfig.TrustSettings, previous.Config)
		} else {
			db.Where("page_type = ?", pageConfig.TrustSettings).Delete(&pageConfig.Entity{})
		}
		hotdataserve.ClearTrustSettingsCache()
	})
	u := users.EntityComplete{Username: "trust-integration", UsernameLower: "trust-integration", CreatedAt: time.Now().Add(-48 * time.Hour), TrustVisitDays: 2, TrustTopicIds: []uint64{1, 2, 3, 4, 5}}
	if err := db.Create(&u).Error; err != nil {
		t.Fatal(err)
	}
	if err := ValidateNewTopic(u.Id); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&posts.Entity{UserId: u.Id, PostNo: 1}).Error; err != nil {
		t.Fatal(err)
	}
	if err := ValidateNewReply(u.Id); err != nil {
		t.Fatal("first post counted as reply", err)
	}
	for i := 0; i < 10; i++ {
		if err := db.Create(&posts.Entity{UserId: u.Id, PostNo: uint64(i + 2)}).Error; err != nil {
			t.Fatal(err)
		}
	}
	var replyQuota *LimitError
	if err := ValidateNewReply(u.Id); !errors.As(err, &replyQuota) || replyQuota.Limit != 10 {
		t.Fatalf("reply quota=%v", err)
	}
	for i := 0; i < 3; i++ {
		if err := db.Create(&topics.Entity{UserId: u.Id, Title: "daily quota draft", Status: 0}).Error; err != nil {
			t.Fatal(err)
		}
	}
	var quota *LimitError
	if err := ValidateNewTopic(u.Id); !errors.As(err, &quota) || quota.Limit != 3 {
		t.Fatalf("daily limit=%v", err)
	}
	config.LevelLimits[0].MaxDailyTopics = new(0)
	raw, _ = json.Marshal(config)
	if err := pageConfig.SaveConfig(pageConfig.TrustSettings, string(raw)); err != nil {
		t.Fatal(err)
	}
	hotdataserve.ClearTrustSettingsCache()
	if err := ValidateNewTopic(u.Id); !errors.As(err, &quota) || quota.Limit != 0 {
		t.Fatalf("zero daily limit=%v", err)
	}
	config.LevelLimits[0].MaxDailyTopics = nil
	raw, _ = json.Marshal(config)
	if err := pageConfig.SaveConfig(pageConfig.TrustSettings, string(raw)); err != nil {
		t.Fatal(err)
	}
	hotdataserve.ClearTrustSettingsCache()
	if err := ValidateNewTopic(u.Id); err != nil {
		t.Fatalf("null daily limit=%v", err)
	}
	if err := ValidateContent(u.Id, "![a](/a.png) ![b](/b.png)", 0); err == nil {
		t.Fatal("image limit bypass")
	}
	if err := Review(u.Id, 0, time.Now()); err != nil {
		t.Fatal(err)
	}
	p, err := GetProgress(u.Id)
	if err != nil || p.Level != trustlevel.Member {
		t.Fatalf("review=%+v %v", p, err)
	}
	if err := SetLevel(u.Id+100, u.Id, trustlevel.Newcomer, false, p.Version, "lower"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("downgrade=%v", err)
	}
	if err := SetLevel(u.Id+100, u.Id, trustlevel.Core, true, p.Version, "appointed"); err != nil {
		t.Fatal(err)
	}
	if err := SetLevel(u.Id+100, u.Id, trustlevel.Core, false, p.Version, "stale"); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale=%v", err)
	}
	rows, err := trusthistory.List(u.Id, 0)
	if err != nil || len(rows) != 2 || rows[0].Reason != "appointed" {
		t.Fatalf("history=%+v %v", rows, err)
	}
	blocked := users.EntityComplete{Username: "trust-restricted", UsernameLower: "trust-restricted", CreatedAt: time.Now().Add(-48 * time.Hour), TrustVisitDays: 2, TrustTopicIds: []uint64{1, 2, 3, 4, 5}, RestrictionStatus: users.RestrictionSuspended}
	if err := db.Create(&blocked).Error; err != nil {
		t.Fatal(err)
	}
	if err := Review(blocked.Id, 0, time.Now()); err != nil {
		t.Fatal(err)
	}
	state, err := users.GetTrustState(blocked.Id)
	if err != nil || state.TrustLevel != trustlevel.Newcomer {
		t.Fatal("suspended account promoted")
	}
	if err := db.Model(&users.EntityComplete{}).Where("id = ?", blocked.Id).Updates(map[string]any{"restriction_status": users.RestrictionNormal, "requires_email_verification": true}).Error; err != nil {
		t.Fatal(err)
	}
	if err := Review(blocked.Id, 0, time.Now()); err != nil {
		t.Fatal(err)
	}
	state, err = users.GetTrustState(blocked.Id)
	if err != nil || state.TrustLevel != trustlevel.Newcomer {
		t.Fatal("unverified account promoted")
	}
	if err := db.Model(&users.EntityComplete{}).Where("id = ?", blocked.Id).Update("is_activated", users.ActivationSuccess).Error; err != nil {
		t.Fatal(err)
	}
	if err := Review(blocked.Id, 0, time.Now()); err != nil {
		t.Fatal(err)
	}
	state, err = users.GetTrustState(blocked.Id)
	if err != nil || state.TrustLevel != trustlevel.Member {
		t.Fatal("verified account failed promotion")
	}
}
