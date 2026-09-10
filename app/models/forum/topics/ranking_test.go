package topics

import (
	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"gorm.io/gorm"
	"testing"
)

func TestHotSortReadsPersistedScoreInAllListPaths(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })
	if err := db.AutoMigrate(&Entity{}, &topicCategoryIndex.Entity{}); err != nil {
		t.Fatal(err)
	}
	// Deliberately omit interaction tables: listing must never calculate scores.
	for id := uint64(1); id <= 260; id++ {
		row := Entity{Id: id, Status: 1, MainCategoryId: 1, ReplyCount: id * 100}
		if err := db.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Exec("INSERT INTO topic_category_index (topic_id, category_id, effective) VALUES (?, 1, 1)", id).Error; err != nil {
			t.Fatal(err)
		}
		if id <= 3 {
			if err := db.Exec("INSERT INTO topic_category_index (topic_id, category_id, effective) VALUES (?, 2, 1)", id).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := db.Table("topics").Where("id IN ?", []uint64{1, 2}).Update("rank_score", 80).Error; err != nil {
		t.Fatal(err)
	}
	for _, categoryID := range []uint64{0, 1, 2} {
		page := PageWithDB(db, PageQuery{Page: 1, PageSize: 10, Sort: "hot", CategoryId: categoryID, FilterStatus: true})
		wantLen, wantNext := 10, true
		if categoryID == 2 {
			wantLen, wantNext = 3, false
		}
		if len(page.Data) != wantLen || page.Data[0].Id != 2 || page.Data[1].Id != 1 || page.HasNext != wantNext {
			t.Fatalf("category %d: unexpected hot page %+v", categoryID, page)
		}
	}
}
