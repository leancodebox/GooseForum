package badgeservice

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/badgeStats"
	"github.com/leancodebox/GooseForum/app/models/forum/badges"
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
	"github.com/leancodebox/GooseForum/app/models/forum/optRecord"
	"github.com/leancodebox/GooseForum/app/models/forum/userBadges"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func badgeAdminTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := dbconnect.Connect()
	models := []any{&badgeStats.Entity{}, &badges.Entity{}, &userBadges.Entity{}, &users.EntityComplete{}, &optRecord.Entity{}, &eventNotification.Entity{}}
	if err := db.Migrator().DropTable(models...); err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatal(err)
	}
	for _, code := range SystemCodes() {
		if err := badgeStats.EnsureWithDB(db, code); err != nil {
			t.Fatal(err)
		}
	}
	InvalidateDefinitions()
	t.Cleanup(InvalidateDefinitions)
	for _, user := range []users.EntityComplete{{Id: 1, Username: "member"}, {Id: 2, Username: "operator"}} {
		if err := db.Create(&user).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func customDefinition() Badge {
	return Badge{Code: "custom_test", Type: badges.TypeCustom, GrantMode: badges.GrantModeManual, Name: "Contributor", IconType: badges.IconTypeAsset, IconURL: "/static/badges/contributor.svg", Color: "blue", Level: LevelBronze, IsEnabled: true, IsWearable: true}
}

func TestNormalizeBadgeRejectsUnsupportedSettings(t *testing.T) {
	for _, test := range []struct {
		name   string
		change func(*Badge)
		field  string
	}{
		{"custom automatic", func(b *Badge) { b.GrantMode = badges.GrantModeAuto }, "grantMode"},
		{"reserved code", func(b *Badge) { b.Code = CodeFirstPost }, "type"},
		{"unknown level", func(b *Badge) { b.Level = "platinum" }, "level"},
		{"unknown color", func(b *Badge) { b.Color = "red" }, "color"},
		{"script url", func(b *Badge) { b.IconURL = "javascript:alert(1)" }, "iconUrl"},
		{"data url", func(b *Badge) { b.IconURL = "data:image/svg+xml,abc" }, "iconUrl"},
		{"protocol relative", func(b *Badge) { b.IconURL = "//example.test/a.png" }, "iconUrl"},
		{"credentials", func(b *Badge) { b.IconURL = "https://user:pass@example.test/a.png" }, "iconUrl"},
		{"backslash", func(b *Badge) { b.IconURL = "/\\example.test/a.png" }, "iconUrl"},
		{"unknown icon type", func(b *Badge) { b.IconType = "html" }, "iconType"},
	} {
		t.Run(test.name, func(t *testing.T) {
			badge := customDefinition()
			test.change(&badge)
			var field *InvalidFieldError
			if err := NormalizeDefinition(&badge); !errors.As(err, &field) || field.Field != test.field {
				t.Fatalf("validation = %v, want field %s", err, test.field)
			}
		})
	}
	for _, url := range []string{"/static/badges/contributor.svg", "/uploads/file.png", "https://cdn.example.test/a.webp?version=3", "http://localhost:5234/a.png"} {
		badge := customDefinition()
		badge.IconURL = url
		if err := NormalizeDefinition(&badge); err != nil {
			t.Fatalf("valid URL %q: %v", url, err)
		}
	}
	badge, _ := systemDefinition(CodeFirstPost)
	badge.GrantMode = badges.GrantModeManual
	if err := NormalizeDefinition(&badge); err != nil || badge.GrantMode != badges.GrantModeAuto {
		t.Fatalf("system rule was changed: %+v %v", badge, err)
	}
}

func TestPersistedAdminCountsAndHistoryDeletionProtection(t *testing.T) {
	badgeAdminTestDB(t)
	if err := SaveDefinition(customDefinition()); err != nil {
		t.Fatal(err)
	}
	lookup := func() AdminBadge {
		t.Helper()
		list, err := ListForAdmin()
		if err != nil {
			t.Fatal(err)
		}
		for _, item := range list {
			if item.Code == "custom_test" {
				return item
			}
		}
		t.Fatal("custom badge missing")
		return AdminBadge{}
	}
	if initial := lookup(); initial.GrantedCount != 0 || !initial.CanDelete {
		t.Fatalf("initial=%+v", initial)
	}
	if _, err := userBadges.GrantManual(1, "custom_test", userBadges.SourceManual, "thanks", 2, nil); err != nil {
		t.Fatal(err)
	}
	if granted := lookup(); granted.GrantedCount != 1 || granted.CanDelete {
		t.Fatalf("cached count did not refresh: %+v", granted)
	}
	if err := DeleteDefinition("custom_test"); !errors.Is(err, ErrBadgeInUse) {
		t.Fatalf("delete granted=%v", err)
	}
	if err := userBadges.Revoke(1, "custom_test"); err != nil {
		t.Fatal(err)
	}
	if revoked := lookup(); revoked.GrantedCount != 0 || revoked.CanDelete {
		t.Fatalf("history lost: %+v", revoked)
	}
	if err := DeleteDefinition("custom_test"); !errors.Is(err, ErrBadgeInUse) {
		t.Fatalf("delete revoked=%v", err)
	}
	if err := DeleteDefinition(CodeFirstPost); !errors.Is(err, ErrSystemBadge) {
		t.Fatalf("delete system=%v", err)
	}
	empty := customDefinition()
	empty.Code = "custom_unused"
	if err := SaveDefinition(empty); err != nil {
		t.Fatal(err)
	}
	if err := DeleteDefinition(empty.Code); err != nil {
		t.Fatal(err)
	}
	if got := ResolveOne(empty.Code); got.Code != "" {
		t.Fatalf("deleted definition cached: %+v", got)
	}
}

func TestRecipientUpdatesValidateAndAudit(t *testing.T) {
	db := badgeAdminTestDB(t)
	if err := SaveDefinition(customDefinition()); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		user                 uint64
		code, action, reason string
	}{
		{99, "custom_test", "grant", "thanks"}, {1, "custom_test", "grant", " "},
		{1, "custom_test", "grant", strings.Repeat("长", 256)}, {1, "custom_test", "invalid", "thanks"},
		{1, CodeFirstPost, "grant", "manual override"},
	} {
		if err := UpdateRecipient(test.code, test.user, test.action, test.reason, 2); err == nil {
			t.Fatalf("accepted invalid update: %+v", test)
		}
	}
	if err := UpdateRecipient("custom_test", 1, "grant", "Helped the community", 2); err != nil {
		t.Fatal(err)
	}
	if err := UpdateRecipient("custom_test", 1, "grant", "duplicate", 2); err != nil {
		t.Fatal(err)
	}
	var logs []optRecord.Entity
	if err := db.Find(&logs).Error; err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 || logs[0].OptUserId != 2 || !strings.Contains(logs[0].OptInfo, "Helped the community") {
		t.Fatalf("grant audit: %+v", logs)
	}
	disabled := customDefinition()
	disabled.IsEnabled = false
	if err := SaveDefinition(disabled); err != nil {
		t.Fatal(err)
	}
	if err := UpdateRecipient("custom_test", 1, "grant", "new award", 2); err == nil {
		t.Fatal("disabled badge granted")
	}
	if got := GetUserBadges(1); len(got) != 0 {
		t.Fatalf("disabled badge shown: %+v", got)
	}
	if err := UpdateRecipient("custom_test", 1, "revoke", "Awarded by mistake", 2); err != nil {
		t.Fatal(err)
	}
	logs = nil
	db.Find(&logs)
	if len(logs) != 2 || !strings.Contains(logs[1].OptInfo, "Awarded by mistake") {
		t.Fatalf("revoke audit: %+v", logs)
	}
	// Automatic awards can also be revoked; later automatic checks must not undo that decision.
	if _, err := userBadges.GrantAuto(1, CodeFirstPost, "first topic", nil); err != nil {
		t.Fatal(err)
	}
	if err := UpdateRecipient(CodeFirstPost, 1, "revoke", "Invalid submission", 2); err != nil {
		t.Fatal(err)
	}
	if changed, err := userBadges.GrantAuto(1, CodeFirstPost, "first topic", nil); err != nil || changed {
		t.Fatalf("automatic regrant = %v %v", changed, err)
	}
}

func TestManualSelectionPreservesDisabledAndAutomaticAwards(t *testing.T) {
	badgeAdminTestDB(t)
	badge := customDefinition()
	badge.IsEnabled = false
	if err := SaveDefinition(badge); err != nil {
		t.Fatal(err)
	}
	if _, err := userBadges.GrantManual(1, badge.Code, userBadges.SourceManual, "earlier", 2, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := userBadges.GrantAuto(1, CodeFirstPost, "first", nil); err != nil {
		t.Fatal(err)
	}
	if err := SaveManualSelection(1, 2, nil); err != nil {
		t.Fatal(err)
	}
	active, err := userBadges.GetActiveByUserIDWithError(1)
	if err != nil || len(active) != 2 {
		t.Fatalf("unrelated awards changed: %+v %v", active, err)
	}
	if err := SaveManualSelection(99, 2, nil); !errors.Is(err, ErrRecipientNotFound) {
		t.Fatalf("missing user=%v", err)
	}
}

func TestAdminRuleMatchesAutomaticThreshold(t *testing.T) {
	rule := RuleForCode(CodeWriter10)
	if rule == nil || rule.Metric != "topics" || rule.Threshold != 10 {
		t.Fatalf("rule=%+v", rule)
	}
	if shouldGrant(CodeWriter10, userStatistics.Entity{TopicCount: 9}) || !shouldGrant(CodeWriter10, userStatistics.Entity{TopicCount: 10}) {
		t.Fatal("grant threshold differs from admin description")
	}
	if RuleForCode(CodeContributor) != nil {
		t.Fatal("manual badge has automatic rule")
	}
}

func TestSystemBadgePreservesClearedPresentationSettings(t *testing.T) {
	badgeAdminTestDB(t)
	badge, _ := systemDefinition(CodeWriter10)
	badge.Description = ""
	badge.SortOrder = 0
	badge.IsEnabled = false
	badge.IsWearable = false
	if err := SaveDefinition(badge); err != nil {
		t.Fatal(err)
	}
	saved := ResolveOne(badge.Code)
	if saved.Description != "" || saved.SortOrder != 0 || saved.IsEnabled || saved.IsWearable {
		t.Fatalf("cleared settings replaced by defaults: %+v", saved)
	}
	if saved.GrantMode != badges.GrantModeAuto {
		t.Fatal("presentation update changed the system grant mode")
	}
}

func TestFirstSystemOverrideCannotInitializeMissingStatistics(t *testing.T) {
	db := badgeAdminTestDB(t)
	badge, _ := systemDefinition(CodeWriter10)
	if granted, err := userBadges.GrantAuto(1, badge.Code, "earned", nil); err != nil || !granted {
		t.Fatalf("grant=%v error=%v", granted, err)
	}
	if err := db.Where("badge_code = ?", badge.Code).Delete(&badgeStats.Entity{}).Error; err != nil {
		t.Fatal(err)
	}
	badge.Name = "Updated appearance"
	if err := SaveDefinition(badge); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("missing statistics should reject the override: %v", err)
	}
	if _, err := badgeStats.Get(badge.Code); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("request initialized missing statistics: %v", err)
	}
	if _, err := badges.GetByCodeWithError(badge.Code); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("failed save persisted an override: %v", err)
	}
	active, err := userBadges.GetActiveByUserIDWithError(1)
	if err != nil || len(active) != 1 || active[0].BadgeCode != badge.Code {
		t.Fatalf("failed save changed existing awards: %+v, %v", active, err)
	}
}

func TestExistingLegacyBadgeCodesRemainEditable(t *testing.T) {
	db := badgeAdminTestDB(t)
	for _, code := range []string{"event.2026", "社区贡献者"} {
		legacy := badges.Entity{Code: code, Type: badges.TypeCustom, GrantMode: badges.GrantModeManual, Name: "Legacy badge", IconType: badges.IconTypeAsset, IconURL: "/static/badges/contributor.svg", Color: "blue", Level: LevelBronze, IsEnabled: true, IsWearable: true}
		if err := db.Create(&legacy).Error; err != nil {
			t.Fatal(err)
		}
		if err := badgeStats.EnsureWithDB(db, code); err != nil {
			t.Fatal(err)
		}
		edited := customDefinition()
		edited.Code = code
		edited.Name = "Renamed badge"
		edited.IsEnabled = false
		if err := SaveDefinition(edited); err != nil {
			t.Fatalf("edit legacy code %q: %v", code, err)
		}
		if saved := ResolveOne(code); saved.Name != edited.Name || saved.IsEnabled {
			t.Fatalf("legacy update did not persist: %+v", saved)
		}
	}
	newBadge := customDefinition()
	newBadge.Code = "new.invalid.code"
	var field *InvalidFieldError
	if err := SaveDefinition(newBadge); !errors.As(err, &field) || field.Field != "code" {
		t.Fatalf("new code validation = %v", err)
	}
}

type badgeSQLRecorder struct {
	logger.Interface
	statements []string
}

func (r *badgeSQLRecorder) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	sql, rows := fc()
	r.statements = append(r.statements, sql)
	r.Interface.Trace(ctx, begin, func() (string, int64) { return sql, rows }, err)
}

func TestBadgeReadsAndDeleteProtectionNeverAggregateGrants(t *testing.T) {
	db := badgeAdminTestDB(t)
	if err := SaveDefinition(customDefinition()); err != nil {
		t.Fatal(err)
	}
	if _, err := userBadges.GrantManual(1, "custom_test", userBadges.SourceManual, "thanks", 2, nil); err != nil {
		t.Fatal(err)
	}
	// Deliberately distinguish the stored value from the grant table contents.
	if err := db.Model(&badgeStats.Entity{}).Where("badge_code = ?", "custom_test").Update("granted_count", 37).Error; err != nil {
		t.Fatal(err)
	}
	previous := db.Logger
	recorder := &badgeSQLRecorder{Interface: previous}
	db.Logger = recorder
	t.Cleanup(func() { db.Logger = previous })
	items, err := ListForAdmin()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range items {
		if item.Code == "custom_test" {
			found = true
			if item.GrantedCount != 37 || item.CanDelete {
				t.Fatalf("stored stats not used: %+v", item)
			}
		}
	}
	if !found {
		t.Fatal("custom badge missing")
	}
	for _, search := range []string{"", "member"} {
		if _, _, err := userBadges.ListRecipients("custom_test", 1, 20, search); err != nil {
			t.Fatal(err)
		}
	}
	if err := DeleteDefinition("custom_test"); !errors.Is(err, ErrBadgeInUse) {
		t.Fatalf("history guard=%v", err)
	}
	if err := db.Where("badge_code = ?", "custom_test").Delete(&badgeStats.Entity{}).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := ListForAdmin(); err == nil {
		t.Fatal("missing stats silently replaced")
	}
	if len(recorder.statements) == 0 {
		t.Fatal("SQL recorder did not run")
	}
	for _, sql := range recorder.statements {
		upper := strings.ToUpper(sql)
		for _, fragment := range []string{"COUNT(", "SUM(", "GROUP BY", "DISTINCT"} {
			if strings.Contains(upper, fragment) {
				t.Fatalf("request aggregated grants: %s", sql)
			}
		}
	}
}

func TestFirstGrantAndDefinitionDeleteCannotLeaveOrphan(t *testing.T) {
	db := badgeAdminTestDB(t)
	for i := range 20 {
		badge := customDefinition()
		badge.Code = fmt.Sprintf("race_%d", i)
		if err := SaveDefinition(badge); err != nil {
			t.Fatal(err)
		}
		var wait sync.WaitGroup
		var granted, deleted bool
		var grantErr, deleteErr error
		wait.Go(func() {
			granted, grantErr = userBadges.GrantManual(1, badge.Code, userBadges.SourceManual, "thanks", 2, nil)
		})
		wait.Go(func() { deleted, deleteErr = badges.DeleteUnusedCustom(badge.Code) })
		wait.Wait()
		if deleteErr != nil {
			t.Fatal(deleteErr)
		}
		if deleted {
			if granted || !errors.Is(grantErr, gorm.ErrRecordNotFound) {
				t.Fatalf("grant survived deleted stats: granted=%v err=%v", granted, grantErr)
			}
			var records []userBadges.Entity
			if err := db.Where("badge_code = ?", badge.Code).Find(&records).Error; err != nil || len(records) != 0 {
				t.Fatalf("orphan=%+v %v", records, err)
			}
		} else {
			if !granted || grantErr != nil {
				t.Fatalf("neither operation succeeded: %v %v", granted, grantErr)
			}
			stat, err := badgeStats.Get(badge.Code)
			if err != nil || stat.GrantedCount != 1 || !stat.HasGrantHistory {
				t.Fatalf("grant stats=%+v %v", stat, err)
			}
		}
	}
}
