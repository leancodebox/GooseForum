package controllers

import (
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/dailyStats"
	"github.com/leancodebox/GooseForum/app/models/forum/fileUsage"
	"github.com/leancodebox/GooseForum/app/models/forum/moderators"
	"github.com/leancodebox/GooseForum/app/models/forum/pointsRecord"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"github.com/leancodebox/GooseForum/app/models/forum/topicUserAction"
	"github.com/leancodebox/GooseForum/app/models/forum/topicUserStat"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/userActivities"
	"github.com/leancodebox/GooseForum/app/models/forum/userBadges"
	"github.com/leancodebox/GooseForum/app/models/forum/userPoints"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"gorm.io/gorm"
)

func setupTopicWriteTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	preferences.Set("db.default.connection", "sqlite")
	preferences.Set("db.default.path", ":memory:")
	conn := dbconnect.Connect()
	err := conn.AutoMigrate(
		&users.EntityComplete{},
		&userStatistics.Entity{},
		&topics.Entity{},
		&posts.Entity{},
		&category.Entity{},
		&topicCategoryIndex.Entity{},
		&topicUserAction.Entity{},
		&topicUserStat.Entity{},
		&fileUsage.Entity{},
		&dailyStats.Entity{},
		&userActivities.Entity{},
		&userPoints.Entity{},
		&pointsRecord.Entity{},
		&userBadges.Entity{},
		&moderators.Entity{},
	)
	if err != nil {
		t.Fatalf("migrate topic write tables: %v", err)
	}
	return conn
}

func createTopicWriteUser(t *testing.T, conn *gorm.DB, id uint64, username string) {
	t.Helper()
	now := time.Now().Add(-time.Hour)
	if err := conn.Create(&users.EntityComplete{Id: id, Username: username, IsActivated: users.ActivationSuccess, CreatedAt: now}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := conn.Create(&userStatistics.Entity{UserId: id}).Error; err != nil {
		t.Fatalf("create user statistics: %v", err)
	}
}

func TestWriteTopicCreatesTopicAndFirstPost(t *testing.T) {
	conn := setupTopicWriteTestDB(t)
	createTopicWriteUser(t, conn, 1001, "author")
	if err := conn.Create(&category.Entity{Id: 2001, Name: "General", Slug: "general"}).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}

	res := WriteTopic(component.BetterRequest[WriteTopicReq]{
		UserId: 1001,
		Params: WriteTopicReq{
			Title:       "Topic title",
			Content:     "Topic content with enough words",
			CategoryId:  []uint64{2001},
			TopicStatus: 1,
		},
	})
	topicID, ok := res.Data.Result.(uint64)
	if !ok || topicID == 0 {
		t.Fatalf("result = %#v, want topic id", res.Data.Result)
	}

	topic := topics.Get(topicID)
	if topic.Id == 0 || topic.Title != "Topic title" || topic.FirstPostId == 0 || topic.PostCount != 1 || topic.PostSeq != 1 {
		t.Fatalf("topic = %#v", topic)
	}
	firstPost := posts.Get(topic.FirstPostId)
	if firstPost.Id == 0 || firstPost.TopicId != topicID || firstPost.PostNo != 1 || firstPost.Content != "Topic content with enough words" {
		t.Fatalf("first post = %#v", firstPost)
	}
	indexes := topicCategoryIndex.GetByTopicId(topicID)
	if len(indexes) != 1 || indexes[0].CategoryId != 2001 || indexes[0].Effective != 1 {
		t.Fatalf("category indexes = %#v", indexes)
	}
}

func TestCreatePostWritesPostAndTopicStats(t *testing.T) {
	conn := setupTopicWriteTestDB(t)
	createTopicWriteUser(t, conn, 1101, "author")
	createTopicWriteUser(t, conn, 1102, "replyer")
	now := time.Now().Add(-time.Hour)
	topic := topics.Entity{Id: 3001, Title: "Topic", UserId: 1101, Status: 1, PostCount: 1, PostSeq: 1, CreatedAt: now, UpdatedAt: now}
	if err := conn.Create(&topic).Error; err != nil {
		t.Fatalf("create topic: %v", err)
	}
	firstPost := posts.Entity{Id: 3101, TopicId: topic.Id, PostNo: 1, UserId: 1101, Content: "first", CreatedAt: now, UpdatedAt: now}
	if err := conn.Create(&firstPost).Error; err != nil {
		t.Fatalf("create first post: %v", err)
	}
	if err := conn.Model(&topics.Entity{}).Where("id = ?", topic.Id).Update("first_post_id", firstPost.Id).Error; err != nil {
		t.Fatalf("set first post: %v", err)
	}

	res := CreatePost(component.BetterRequest[CreatePostReq]{
		UserId: 1102,
		Params: CreatePostReq{
			TopicId: topic.Id,
			Content: "reply content with enough words",
		},
	})
	payload, ok := res.Data.Result.(map[string]any)
	if !ok {
		t.Fatalf("result = %#v, want payload", res.Data.Result)
	}
	postID, ok := payload["id"].(uint64)
	if !ok || postID == 0 {
		t.Fatalf("reply payload = %#v", payload)
	}
	post := posts.Get(postID)
	if post.Id == 0 || post.TopicId != topic.Id || post.PostNo != 2 {
		t.Fatalf("post = %#v", post)
	}
	topic = topics.Get(topic.Id)
	if topic.PostCount != 2 || topic.ReplyCount != 1 || topic.PostSeq != 2 {
		t.Fatalf("topic stats = %#v", topic)
	}
}

func TestTopicActionsUseTopicUserAction(t *testing.T) {
	conn := setupTopicWriteTestDB(t)
	createTopicWriteUser(t, conn, 1201, "author")
	createTopicWriteUser(t, conn, 1202, "reader")
	if err := conn.Create(&topics.Entity{Id: 4001, Title: "Topic", UserId: 1201, Status: 1}).Error; err != nil {
		t.Fatalf("create topic: %v", err)
	}

	LikeTopic(component.BetterRequest[LikeTopicReq]{UserId: 1202, Params: LikeTopicReq{TopicId: 4001, Action: 1}})
	BookmarkTopic(component.BetterRequest[BookmarkTopicReq]{UserId: 1202, Params: BookmarkTopicReq{TopicId: 4001, Action: 1}})
	WatchTopic(component.BetterRequest[WatchTopicReq]{UserId: 1202, Params: WatchTopicReq{TopicId: 4001, Action: 1}})

	action := topicUserAction.GetByTopicId(uint64(1202), uint64(4001))
	if action.Id == 0 || action.LikedAt == nil || action.BookmarkedAt == nil || action.WatchedAt == nil {
		t.Fatalf("topic action = %#v", action)
	}
	topic := topics.Get(4001)
	if topic.LikeCount != 1 {
		t.Fatalf("like count = %d, want 1", topic.LikeCount)
	}
}
