package postservice

import (
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topicUserStat"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
)

func TestCreateTopicPostRollsBackSequenceAndStatsWhenPostInsertFails(t *testing.T) {
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&topics.Entity{}, &posts.Entity{}, &topicUserStat.Entity{}); err != nil {
		t.Fatalf("migrate reply tables: %v", err)
	}
	const topicID uint64 = 98000001
	const conflictingPostID uint64 = 98000002
	conn.Unscoped().Where("topic_id = ?", topicID).Delete(&topicUserStat.Entity{})
	conn.Unscoped().Where("topic_id = ? OR id = ?", topicID, conflictingPostID).Delete(&posts.Entity{})
	conn.Unscoped().Where("id = ?", topicID).Delete(&topics.Entity{})

	now := time.Now().Add(-time.Hour)
	topic := topics.Entity{Id: topicID, Title: "atomic reply", UserId: 7, Status: 1, PostSeq: 1, PostCount: 1, LastPostedAt: &now}
	if err := conn.Create(&topic).Error; err != nil {
		t.Fatalf("create topic: %v", err)
	}
	if err := conn.Create(&posts.Entity{Id: conflictingPostID, TopicId: topicID + 1, PostNo: 1, UserId: 8, Content: "conflict"}).Error; err != nil {
		t.Fatalf("create conflicting post: %v", err)
	}

	reply := posts.Entity{Id: conflictingPostID, TopicId: topicID, UserId: 9, Content: "reply"}
	if err := CreateTopicPost(&reply, topic); err == nil {
		t.Fatal("CreateTopicPost error = nil, want duplicate primary-key failure")
	}
	stored := topics.GetSimple(topicID)
	if stored.PostSeq != 1 || stored.PostCount != 1 || stored.ReplyCount != 0 {
		t.Fatalf("topic stats after rollback = seq:%d posts:%d replies:%d", stored.PostSeq, stored.PostCount, stored.ReplyCount)
	}
	var statCount int64
	if err := conn.Model(&topicUserStat.Entity{}).Where("topic_id = ?", topicID).Count(&statCount).Error; err != nil {
		t.Fatalf("count reply stats: %v", err)
	}
	if statCount != 0 {
		t.Fatalf("reply stats after rollback = %d, want 0", statCount)
	}
}
