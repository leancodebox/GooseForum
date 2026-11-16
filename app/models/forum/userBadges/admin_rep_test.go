package userBadges_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/badgeStats"
	"github.com/leancodebox/GooseForum/app/models/forum/badges"
	records "github.com/leancodebox/GooseForum/app/models/forum/userBadges"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/datamigration"
	"gorm.io/gorm"
)

func recipientTestDB(t *testing.T) (*gorm.DB, string) {
	t.Helper()
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}, &records.Entity{}, &badges.Entity{}, &badgeStats.Entity{}); err != nil {
		t.Fatal(err)
	}
	code := fmt.Sprintf("test_%d", time.Now().UnixNano())
	if err := badgeStats.EnsureWithDB(db, code); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("badge_code = ?", code).Delete(&records.Entity{})
		db.Where("badge_code = ?", code).Delete(&badgeStats.Entity{})
		db.Unscoped().Where("username LIKE ?", code+"%").Delete(&users.EntityComplete{})
	})
	return db, code
}

func recipientTestUser(t *testing.T, db *gorm.DB, username string) users.EntityComplete {
	t.Helper()
	user := users.EntityComplete{Username: username, AvatarUrl: "/avatar.png"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	return user
}

func TestRecipientCountsSearchAndHistoryRespectRevocationAndDeletedUsers(t *testing.T) {
	db, code := recipientTestDB(t)
	actor := recipientTestUser(t, db, code+"-admin")
	alice := recipientTestUser(t, db, code+"-Alice")
	bob := recipientTestUser(t, db, code+"-Bob")
	revoked := recipientTestUser(t, db, code+"-revoked")
	deleted := recipientTestUser(t, db, code+"-deleted")
	now := time.Now().UTC().Truncate(time.Second)
	for _, user := range []users.EntityComplete{alice, bob, revoked, deleted} {
		row := records.Entity{UserId: user.Id, BadgeCode: code, Source: records.SourceManual, Reason: "helpful", GrantedBy: actor.Id, GrantedAt: now}
		if user.Id == revoked.Id {
			row.RevokedAt = &now
		}
		if err := db.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Delete(&deleted).Error; err != nil {
		t.Fatal(err)
	}
	if err := datamigration.BackfillBadgeStatsWithDB(db, []string{code}); err != nil {
		t.Fatal(err)
	}
	counts, err := badgeStats.GetByCodes([]string{code, "absent"})
	if err != nil || counts[code].GrantedCount != 2 || counts["absent"].GrantedCount != 0 {
		t.Fatalf("counts=%v err=%v", counts, err)
	}
	rows, hasNext, err := records.ListRecipients(code, 1, 1, "")
	if err != nil || !hasNext || len(rows) != 1 || rows[0].UserId != bob.Id || rows[0].GrantedByName != actor.Username {
		t.Fatalf("first page=%+v hasNext=%v err=%v", rows, hasNext, err)
	}
	rows, hasNext, err = records.ListRecipients(code, 2, 1, "")
	if err != nil || hasNext || len(rows) != 1 || rows[0].UserId != alice.Id {
		t.Fatalf("second page=%+v hasNext=%v err=%v", rows, hasNext, err)
	}
	for _, search := range []string{"ALICE", fmt.Sprint(alice.Id)} {
		rows, hasNext, err = records.ListRecipients(code, 1, 20, search)
		if err != nil || hasNext || len(rows) != 1 || rows[0].UserId != alice.Id {
			t.Fatalf("search=%q rows=%+v hasNext=%v err=%v", search, rows, hasNext, err)
		}
	}
	if err := db.Delete(&actor).Error; err != nil {
		t.Fatal(err)
	}
	rows, _, err = records.ListRecipients(code, 1, 20, "")
	if err != nil || len(rows) != 2 || rows[0].GrantedByName != "" || rows[0].GrantedBy != actor.Id {
		t.Fatalf("deleted grantor rows=%+v err=%v", rows, err)
	}
	for _, user := range []users.EntityComplete{alice, bob, revoked, deleted} {
		if err := records.Revoke(user.Id, code); err != nil {
			t.Fatal(err)
		}
	}
	counts, err = badgeStats.GetByCodes([]string{code})
	if err != nil || counts[code].GrantedCount != 0 {
		t.Fatalf("revoked counts=%v err=%v", counts, err)
	}
	history, err := badgeStats.GetByCodes([]string{code, "absent"})
	if err != nil || !history[code].HasGrantHistory || history["absent"].HasGrantHistory {
		t.Fatalf("history=%v err=%v", history, err)
	}
}

func TestRecipientUpdatesAreIdempotentAndRestoreOnlyRevokedGrants(t *testing.T) {
	db, code := recipientTestDB(t)
	user := recipientTestUser(t, db, code+"-user")
	changed, err := records.GrantManual(user.Id, code, records.SourceManual, "original", 7, records.Metadata{"event": "welcome"})
	if err != nil || !changed {
		t.Fatalf("initial grant changed=%v err=%v", changed, err)
	}
	before, err := records.GetActiveByUserIDWithError(user.Id)
	if err != nil || len(before) != 1 {
		t.Fatalf("before=%+v err=%v", before, err)
	}
	changed, err = records.UpdateRecipient(user.Id, code, "grant", "retry", 8)
	if err != nil || changed {
		t.Fatalf("repeat grant changed=%v err=%v", changed, err)
	}
	after, err := records.GetActiveByUserIDWithError(user.Id)
	if err != nil || len(after) != 1 || after[0].Reason != "original" || after[0].GrantedBy != 7 || !after[0].GrantedAt.Equal(before[0].GrantedAt) || after[0].Metadata["event"] != "welcome" {
		t.Fatalf("repeat grant overwrote original: %+v err=%v", after, err)
	}
	for i := range 2 {
		changed, err = records.UpdateRecipient(user.Id, code, "revoke", "mistake", 9)
		if err != nil || changed != (i == 0) {
			t.Fatalf("revoke attempt %d changed=%v err=%v", i, changed, err)
		}
	}
	if changed, err = records.GrantAuto(user.Id, code, "auto retry", nil); err != nil || changed {
		t.Fatalf("automatic grant restored revoked badge: changed=%v err=%v", changed, err)
	}
	changed, err = records.GrantManual(user.Id, code, records.SourceManual, "restored", 10, records.Metadata{"event": "review"})
	if err != nil || !changed {
		t.Fatalf("restore changed=%v err=%v", changed, err)
	}
	after, err = records.GetActiveByUserIDWithError(user.Id)
	if err != nil || len(after) != 1 || after[0].Reason != "restored" || after[0].GrantedBy != 10 || after[0].Metadata["event"] != "review" {
		t.Fatalf("restore=%+v err=%v", after, err)
	}
}

func TestConcurrentRecipientGrantsHaveOneWinner(t *testing.T) {
	db, code := recipientTestDB(t)
	user := recipientTestUser(t, db, code+"-user")
	for _, restore := range []bool{false, true} {
		if restore {
			if err := records.Revoke(user.Id, code); err != nil {
				t.Fatal(err)
			}
		}
		var wait sync.WaitGroup
		results := make(chan bool, 12)
		failures := make(chan error, 12)
		for range 12 {
			wait.Go(func() {
				changed, err := records.UpdateRecipient(user.Id, code, "grant", "concurrent", 7)
				results <- changed
				failures <- err
			})
		}
		wait.Wait()
		close(results)
		close(failures)
		winners := 0
		for changed := range results {
			if changed {
				winners++
			}
		}
		for err := range failures {
			if err != nil {
				t.Fatal(err)
			}
		}
		if winners != 1 {
			t.Fatalf("restore=%v winners=%d", restore, winners)
		}
	}
}

func TestRecipientReadsReturnDatabaseErrors(t *testing.T) {
	db, _ := recipientTestDB(t)
	const callback = "test:badge_query_failure"
	errUnavailable := errors.New("database unavailable")
	if err := db.Callback().Query().Before("gorm:query").Register(callback, func(tx *gorm.DB) { tx.AddError(errUnavailable) }); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Callback().Query().Remove(callback) })
	if _, err := records.GetActiveByUserIDWithError(1); !errors.Is(err, errUnavailable) {
		t.Fatalf("read error=%v", err)
	}
}
