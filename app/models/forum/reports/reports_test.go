package reports

import (
	"sync"
	"testing"

	db "github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"gorm.io/gorm/schema"
)

func TestReportEntityDoesNotExposeLegacySourceID(t *testing.T) {
	parsed, err := schema.Parse(&Entity{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse reports entity schema: %v", err)
	}
	legacyFieldName := "Art" + "icleId"
	legacyDBName := "art" + "icle_id"
	if _, ok := parsed.FieldsByName[legacyFieldName]; ok {
		t.Fatalf("reports.Entity should not expose legacy source id field")
	}
	if _, ok := parsed.FieldsByDBName[legacyDBName]; ok {
		t.Fatalf("reports.Entity should not map legacy source id column")
	}
}

func TestCreateOpenReturnsExistingOpenReport(t *testing.T) {
	preferences.Set("db.default.connection", "sqlite")
	preferences.Set("db.default.path", ":memory:")
	if err := db.Connect().AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("migrate reports: %v", err)
	}

	first, created, err := CreateOpen(Entity{
		TargetType: TargetTopic,
		TargetId:   1,
		ReporterId: 2,
		Reason:     ReasonSpam,
	})
	if err != nil || !created {
		t.Fatalf("first CreateOpen() id=%d created=%v err=%v, want created", first.Id, created, err)
	}

	second, created, err := CreateOpen(Entity{
		TargetType: TargetTopic,
		TargetId:   1,
		ReporterId: 2,
		Reason:     ReasonAbuse,
	})
	if err != nil {
		t.Fatalf("second CreateOpen() err=%v", err)
	}
	if created {
		t.Fatal("second CreateOpen() created duplicate open report")
	}
	if second.Id != first.Id || second.Reason != ReasonSpam {
		t.Fatalf("second CreateOpen() = id %d reason %q, want existing id %d reason %q", second.Id, second.Reason, first.Id, first.Reason)
	}
}

func TestUpdateStatusRecordsHandler(t *testing.T) {
	preferences.Set("db.default.connection", "sqlite")
	preferences.Set("db.default.path", ":memory:")
	if err := db.Connect().AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("migrate reports: %v", err)
	}

	report, created, err := CreateOpen(Entity{
		TargetType: TargetPost,
		TargetId:   9,
		ReporterId: 2,
		Reason:     ReasonAbuse,
	})
	if err != nil || !created {
		t.Fatalf("CreateOpen() id=%d created=%v err=%v, want created", report.Id, created, err)
	}

	if err := UpdateStatus(report.Id, StatusRejected, ResolutionIgnored, 7); err != nil {
		t.Fatalf("UpdateStatus() err=%v", err)
	}

	updated := Get(report.Id)
	if updated.HandlerId != 7 || updated.HandledAt == nil {
		t.Fatalf("handler id=%d handledAt=%v, want handler 7 and handledAt", updated.HandlerId, updated.HandledAt)
	}
}

func TestCursorPageScopeCategoryIDsIncludesTopicsAndPosts(t *testing.T) {
	preferences.Set("db.default.connection", "sqlite")
	preferences.Set("db.default.path", ":memory:")
	conn := db.Connect()
	if err := conn.AutoMigrate(&Entity{}, &topicCategoryIndex.Entity{}); err != nil {
		t.Fatalf("migrate reports scope tables: %v", err)
	}
	conn.Where("1 = 1").Delete(&Entity{})
	conn.Where("1 = 1").Delete(&topicCategoryIndex.Entity{})

	conn.Create(&topicCategoryIndex.Entity{TopicId: 10, CategoryId: 3, Effective: 1})
	conn.Create(&topicCategoryIndex.Entity{TopicId: 20, CategoryId: 4, Effective: 1})
	conn.Create(&[]Entity{
		{TargetType: TargetTopic, TargetId: 10, TopicId: 10, ReporterId: 1, Status: StatusOpen},
		{TargetType: TargetTopic, TargetId: 20, TopicId: 20, ReporterId: 1, Status: StatusOpen},
		{TargetType: TargetPost, TargetId: 100, TopicId: 10, ReporterId: 1, Status: StatusOpen},
		{TargetType: TargetPost, TargetId: 200, TopicId: 20, ReporterId: 1, Status: StatusOpen},
	})

	got := CursorPage(CursorPageQuery{
		Status:           StatusOpen,
		ScopeCategoryIDs: []uint64{3},
		PageSize:         10,
	})
	if len(got) != 2 {
		t.Fatalf("CursorPage scoped len=%d, want 2: %#v", len(got), got)
	}
	for _, row := range got {
		if row.TargetId != 10 && row.TargetId != 100 {
			t.Fatalf("CursorPage scoped returned target %s/%d outside category 3", row.TargetType, row.TargetId)
		}
	}
}
