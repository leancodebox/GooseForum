package userActivities

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
)

func TestGetUserTimelineForAudienceFiltersBeforePagination(t *testing.T) {
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&topics.Entity{}, &topicCategoryIndex.Entity{}, &Entity{}); err != nil {
		t.Fatalf("migrate timeline tables: %v", err)
	}
	userID := uint64(987001)
	topicIDs := []uint64{987010, 987020}
	activityIDs := []uint64{987101, 987102, 987103}
	t.Cleanup(func() {
		conn.Delete(&Entity{}, "id IN ?", activityIDs)
		conn.Delete(&topicCategoryIndex.Entity{}, "topic_id IN ?", topicIDs)
		conn.Unscoped().Delete(&topics.Entity{}, "id IN ?", topicIDs)
	})
	conn.Delete(&Entity{}, "id IN ?", activityIDs)
	conn.Delete(&topicCategoryIndex.Entity{}, "topic_id IN ?", topicIDs)
	conn.Unscoped().Delete(&topics.Entity{}, "id IN ?", topicIDs)
	if err := conn.Create(&[]topics.Entity{
		{Id: topicIDs[0], Title: "public", CategoryIds: []uint64{9871}, MainCategoryId: 9871, Status: 1},
		{Id: topicIDs[1], Title: "private", CategoryIds: []uint64{9872}, MainCategoryId: 9872, Status: 1},
	}).Error; err != nil {
		t.Fatalf("create topics: %v", err)
	}
	if err := conn.Create(&[]topicCategoryIndex.Entity{
		{TopicId: topicIDs[0], CategoryId: 9871, Effective: 1},
		{TopicId: topicIDs[1], CategoryId: 9872, Effective: 1},
	}).Error; err != nil {
		t.Fatalf("create topic category indexes: %v", err)
	}
	if err := conn.Create(&[]Entity{
		{Id: activityIDs[0], UserId: userID, TopicId: topicIDs[0], Action: int(ActionPost), SubjectType: SubjectTopic, SubjectId: topicIDs[0]},
		{Id: activityIDs[1], UserId: userID, TopicId: topicIDs[1], Action: int(ActionPost), SubjectType: SubjectTopic, SubjectId: topicIDs[1]},
		{Id: activityIDs[2], UserId: userID, TopicId: 0, Action: int(ActionFollow), SubjectType: SubjectUser, SubjectId: 1},
	}).Error; err != nil {
		t.Fatalf("create activities: %v", err)
	}

	rows, err := GetUserTimelineForAudience(userID, 0, 2, []uint64{9871}, true)
	if err != nil {
		t.Fatalf("GetUserTimelineForAudience: %v", err)
	}
	if len(rows) != 2 || rows[0].Id != activityIDs[2] || rows[1].Id != activityIDs[0] {
		t.Fatalf("filtered timeline ids = %#v", rows)
	}
}
