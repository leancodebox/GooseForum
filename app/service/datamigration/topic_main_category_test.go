package datamigration

import (
	"errors"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/service/searchservice"
	"gorm.io/gorm"
)

func TestBackfillTopicMainCategoryUsesFirstNonZeroCategoryAndIsRerunnable(t *testing.T) {
	conn, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := conn.AutoMigrate(&topics.Entity{}); err != nil {
		t.Fatalf("migrate topics: %v", err)
	}
	originalUpdatedAt := time.Date(2025, 2, 3, 4, 5, 6, 0, time.UTC)
	rows := []topics.Entity{
		{Id: 1, Title: "first", CategoryIds: []uint64{0, 7, 8}, UpdatedAt: originalUpdatedAt},
		{Id: 2, Title: "existing", CategoryIds: []uint64{9}, MainCategoryId: 10, UpdatedAt: originalUpdatedAt},
		{Id: 3, Title: "orphan", CategoryIds: []uint64{}, UpdatedAt: originalUpdatedAt},
	}
	if err := conn.Create(&rows).Error; err != nil {
		t.Fatalf("create topics: %v", err)
	}

	result := BackfillTopicMainCategoryWithDB(conn)
	if result.Failed != 0 || result.Updated != 1 || result.Orphaned != 1 {
		t.Fatalf("unexpected first migration result: %#v", result)
	}
	var migrated []topics.Entity
	if err := conn.Order("id").Find(&migrated).Error; err != nil {
		t.Fatalf("reload topics: %v", err)
	}
	if migrated[0].MainCategoryId != 7 || migrated[1].MainCategoryId != 10 || migrated[2].MainCategoryId != 0 {
		t.Fatalf("unexpected main categories: %d, %d, %d", migrated[0].MainCategoryId, migrated[1].MainCategoryId, migrated[2].MainCategoryId)
	}
	if !migrated[0].UpdatedAt.Equal(originalUpdatedAt) {
		t.Fatalf("migration changed topic updated_at: got %s want %s", migrated[0].UpdatedAt, originalUpdatedAt)
	}

	rerun := BackfillTopicMainCategoryWithDB(conn)
	if rerun.Failed != 0 || rerun.Updated != 0 || rerun.Orphaned != 1 {
		t.Fatalf("unexpected rerun result: %#v", rerun)
	}
}

func TestMainCategorySearchMigrationSkipsOnlyWhenNotConfigured(t *testing.T) {
	buildCalled := false
	result := rebuildTopicMainCategorySearchIndex(
		func() bool { return false },
		func() bool { return false },
		func() (*searchservice.IndexBuildResult, error) {
			buildCalled = true
			return nil, nil
		},
	)
	if !result.Skipped || result.Failed != 0 || buildCalled {
		t.Fatalf("unexpected unconfigured result: %#v, buildCalled=%v", result, buildCalled)
	}

	result = rebuildTopicMainCategorySearchIndex(
		func() bool { return true },
		func() bool { return false },
		func() (*searchservice.IndexBuildResult, error) {
			t.Fatal("build should not run while Meilisearch is unavailable")
			return nil, nil
		},
	)
	if result.Skipped || result.Failed != 1 || result.LastFailed == "" {
		t.Fatalf("unexpected unavailable result: %#v", result)
	}
}

func TestMainCategorySearchMigrationReportsBuildFailures(t *testing.T) {
	result := rebuildTopicMainCategorySearchIndex(
		func() bool { return true },
		func() bool { return true },
		func() (*searchservice.IndexBuildResult, error) { return nil, errors.New("boom") },
	)
	if result.Failed != 1 || result.LastFailed != "boom" {
		t.Fatalf("unexpected build error result: %#v", result)
	}

	result = rebuildTopicMainCategorySearchIndex(
		func() bool { return true },
		func() bool { return true },
		func() (*searchservice.IndexBuildResult, error) { return nil, nil },
	)
	if result.Failed != 1 || result.LastFailed == "" {
		t.Fatalf("unexpected nil build result: %#v", result)
	}

	result = rebuildTopicMainCategorySearchIndex(
		func() bool { return true },
		func() bool { return true },
		func() (*searchservice.IndexBuildResult, error) {
			return &searchservice.IndexBuildResult{ProcessedCount: 8, FailedCount: 2}, nil
		},
	)
	if !result.Rebuilt || result.ProcessedCount != 8 || result.FailedCount != 2 || result.Failed != 1 {
		t.Fatalf("unexpected partial rebuild result: %#v", result)
	}
}
