package datamigration

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/userActivities"
	"gorm.io/gorm"
)

func TestBackfillUserActivityTopicIDs(t *testing.T) {
	conn, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := conn.AutoMigrate(&posts.Entity{}, &userActivities.Entity{}); err != nil {
		t.Fatalf("migrate activity tables: %v", err)
	}
	if err := conn.Create(&posts.Entity{Id: 501, TopicId: 101, PostNo: 2}).Error; err != nil {
		t.Fatalf("create post: %v", err)
	}
	rows := []userActivities.Entity{
		{Id: 1, UserId: 1, SubjectType: userActivities.SubjectTopic, SubjectId: 100},
		{Id: 2, UserId: 1, SubjectType: userActivities.SubjectPost, SubjectId: 501},
		{Id: 3, UserId: 1, SubjectType: userActivities.SubjectPost, SubjectId: 999},
		{Id: 4, UserId: 1, SubjectType: userActivities.SubjectUser, SubjectId: 2},
	}
	if err := conn.Create(&rows).Error; err != nil {
		t.Fatalf("create activities: %v", err)
	}

	result := BackfillUserActivityTopicIDsWithDB(conn)
	if result.Failed != 0 || result.Topics != 1 || result.Missing != 1 {
		t.Fatalf("migration result = %+v", result)
	}
	var stored []userActivities.Entity
	if err := conn.Order("id ASC").Find(&stored).Error; err != nil {
		t.Fatalf("load activities: %v", err)
	}
	want := []uint64{100, 101, 0, 0}
	for i := range stored {
		if stored[i].TopicId != want[i] {
			t.Fatalf("activity %d topic_id = %d, want %d", stored[i].Id, stored[i].TopicId, want[i])
		}
	}
}
