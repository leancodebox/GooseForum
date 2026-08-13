package datamigration

import (
	"reflect"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/accessGroups"
	"github.com/leancodebox/GooseForum/app/models/forum/categoryGroupPermissions"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"gorm.io/gorm"
)

func TestEnforceSingleRestrictedTopicCategoryPreservesAudience(t *testing.T) {
	conn, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := conn.AutoMigrate(
		&accessGroups.Entity{}, &categoryGroupPermissions.Entity{},
		&topics.Entity{}, &topicCategoryIndex.Entity{},
	); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}
	everyoneKey := accessGroups.SystemKeyEveryone
	everyone := accessGroups.Entity{Id: 1, Name: "Everyone", SystemKey: &everyoneKey, JoinMode: accessGroups.JoinModeSystem, Status: accessGroups.StatusEnabled}
	if err := conn.Create(&everyone).Error; err != nil {
		t.Fatalf("create everyone: %v", err)
	}
	for _, categoryID := range []uint64{1, 2} {
		grant := categoryGroupPermissions.Entity{CategoryId: categoryID, AccessGroupId: everyone.Id, PermissionLevel: categoryGroupPermissions.PermissionRead, Status: categoryGroupPermissions.StatusEnabled}
		if err := conn.Create(&grant).Error; err != nil {
			t.Fatalf("create public grant: %v", err)
		}
	}
	rows := []topics.Entity{
		{Id: 1, Title: "public main", CategoryIds: []uint64{1, 3, 2}, MainCategoryId: 1},
		{Id: 2, Title: "restricted main", CategoryIds: []uint64{3, 1}, MainCategoryId: 3},
		{Id: 3, Title: "all public", CategoryIds: []uint64{1, 2}, MainCategoryId: 1},
		{Id: 4, Title: "single restricted", CategoryIds: []uint64{3}, MainCategoryId: 3},
		{Id: 5, Title: "deleted restricted main", CategoryIds: []uint64{3, 2}, MainCategoryId: 3},
	}
	if err := conn.Create(&rows).Error; err != nil {
		t.Fatalf("create topics: %v", err)
	}
	for _, row := range rows {
		for _, categoryID := range row.CategoryIds {
			if err := conn.Create(&topicCategoryIndex.Entity{TopicId: row.Id, CategoryId: categoryID, Effective: 1}).Error; err != nil {
				t.Fatalf("create category index: %v", err)
			}
		}
	}
	if err := conn.Delete(&topics.Entity{}, 5).Error; err != nil {
		t.Fatalf("soft delete legacy topic: %v", err)
	}

	result := EnforceSingleRestrictedTopicCategoryWithDB(conn)
	if result.Failed != 0 || result.Updated != 3 || result.RemovedCategoryLinks != 3 {
		t.Fatalf("migration result = %#v", result)
	}
	want := map[uint64][]uint64{1: {1, 2}, 2: {3}, 3: {1, 2}, 4: {3}, 5: {3}}
	for topicID, categoryIDs := range want {
		var row topics.Entity
		if err := conn.Unscoped().First(&row, topicID).Error; err != nil {
			t.Fatalf("load topic %d: %v", topicID, err)
		}
		if !reflect.DeepEqual(row.CategoryIds, categoryIDs) || row.MainCategoryId != categoryIDs[0] {
			t.Fatalf("topic %d categories=%v main=%d want=%v", topicID, row.CategoryIds, row.MainCategoryId, categoryIDs)
		}
		var indexes []topicCategoryIndex.Entity
		if err := conn.Where("topic_id = ? AND effective = ?", topicID, 1).Order("category_id").Find(&indexes).Error; err != nil {
			t.Fatalf("load topic %d indexes: %v", topicID, err)
		}
		gotIDs := make([]uint64, 0, len(indexes))
		for _, index := range indexes {
			gotIDs = append(gotIDs, index.CategoryId)
		}
		orderedWant := append([]uint64(nil), categoryIDs...)
		if len(orderedWant) > 1 && orderedWant[0] > orderedWant[1] {
			orderedWant[0], orderedWant[1] = orderedWant[1], orderedWant[0]
		}
		if !reflect.DeepEqual(gotIDs, orderedWant) {
			t.Fatalf("topic %d active indexes=%v want=%v", topicID, gotIDs, orderedWant)
		}
	}
	rerun := EnforceSingleRestrictedTopicCategoryWithDB(conn)
	if rerun.Failed != 0 || rerun.Updated != 0 || rerun.RemovedCategoryLinks != 0 {
		t.Fatalf("rerun result = %#v", rerun)
	}
}

func TestEnforceSingleRestrictedTopicCategoryRollsBackBatchAndCounters(t *testing.T) {
	conn, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := conn.AutoMigrate(
		&accessGroups.Entity{}, &categoryGroupPermissions.Entity{},
		&topics.Entity{}, &topicCategoryIndex.Entity{},
	); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}
	everyoneKey := accessGroups.SystemKeyEveryone
	if err := conn.Create(&accessGroups.Entity{
		Id: 1, Name: "Everyone", SystemKey: &everyoneKey,
		JoinMode: accessGroups.JoinModeSystem, Status: accessGroups.StatusEnabled,
	}).Error; err != nil {
		t.Fatalf("create everyone: %v", err)
	}
	if err := conn.Create(&categoryGroupPermissions.Entity{
		CategoryId: 1, AccessGroupId: 1,
		PermissionLevel: categoryGroupPermissions.PermissionRead,
		Status:          categoryGroupPermissions.StatusEnabled,
	}).Error; err != nil {
		t.Fatalf("create public grant: %v", err)
	}
	rows := []topics.Entity{
		{Id: 1, Title: "first", CategoryIds: []uint64{1, 3}, MainCategoryId: 1},
		{Id: 2, Title: "second", CategoryIds: []uint64{3, 1}, MainCategoryId: 3},
	}
	if err := conn.Create(&rows).Error; err != nil {
		t.Fatalf("create topics: %v", err)
	}
	for _, row := range rows {
		for _, categoryID := range row.CategoryIds {
			if err := conn.Create(&topicCategoryIndex.Entity{TopicId: row.Id, CategoryId: categoryID, Effective: 1}).Error; err != nil {
				t.Fatalf("create category index: %v", err)
			}
		}
	}
	if err := conn.Exec(`CREATE TRIGGER fail_restricted_category_migration
		BEFORE UPDATE ON topics WHEN NEW.id = 2
		BEGIN SELECT RAISE(ABORT, 'forced migration failure'); END`).Error; err != nil {
		t.Fatalf("create failure trigger: %v", err)
	}

	result := EnforceSingleRestrictedTopicCategoryWithDB(conn)
	if result.Failed != 1 || result.Updated != 0 || result.RemovedCategoryLinks != 0 {
		t.Fatalf("failed migration result = %#v", result)
	}
	var first topics.Entity
	if err := conn.First(&first, 1).Error; err != nil {
		t.Fatalf("load rolled back topic: %v", err)
	}
	if !reflect.DeepEqual(first.CategoryIds, []uint64{1, 3}) {
		t.Fatalf("first topic was not rolled back: %v", first.CategoryIds)
	}
	var activeIndexCount int64
	if err := conn.Model(&topicCategoryIndex.Entity{}).
		Where("topic_id = ? AND effective = ?", 1, 1).
		Count(&activeIndexCount).Error; err != nil {
		t.Fatalf("count rolled back indexes: %v", err)
	}
	if activeIndexCount != 2 {
		t.Fatalf("active indexes after rollback = %d", activeIndexCount)
	}
}
