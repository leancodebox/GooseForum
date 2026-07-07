package datamigration

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
	"github.com/leancodebox/GooseForum/app/models/forum/migrationMapping"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/reports"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"github.com/leancodebox/GooseForum/app/models/forum/topicUserAction"
	"github.com/leancodebox/GooseForum/app/models/forum/topicUserStat"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"gorm.io/gorm"
)

func TestBackfillTopicPostModelCopiesOldTablesWithoutOldModels(t *testing.T) {
	conn, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	createLegacyTopicPostTables(t, conn)
	if err := conn.AutoMigrate(
		&topics.Entity{},
		&posts.Entity{},
		&category.Entity{},
		&topicCategoryIndex.Entity{},
		&topicUserAction.Entity{},
		&topicUserStat.Entity{},
		&migrationMapping.Entity{},
		&eventNotification.Entity{},
		&reports.Entity{},
	); err != nil {
		t.Fatalf("migrate new tables: %v", err)
	}

	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	conn.Table("articles").Create(map[string]any{
		"id":               10,
		"title":            "hello topic",
		"content":          "first post",
		"description":      "summary",
		"first_image_url":  "/a.png",
		"rendered_html":    "<p>first post</p>",
		"rendered_version": 2,
		"category_id":      `[3]`,
		"user_id":          1,
		"posters":          `[{"user_id":1,"description":"author"}]`,
		"article_status":   1,
		"process_status":   0,
		"like_count":       5,
		"view_count":       8,
		"reply_count":      1,
		"reply_seq":        1,
		"pin_weight":       7,
		"created_at":       now,
		"updated_at":       now,
	})
	conn.Table("reply").Create(map[string]any{
		"id":               99,
		"article_id":       10,
		"reply_no":         1,
		"user_id":          2,
		"target_id":        0,
		"content":          "reply body",
		"rendered_html":    "<p>reply body</p>",
		"rendered_version": 3,
		"process_status":   0,
		"reply_id":         0,
		"created_at":       now.Add(time.Minute),
		"updated_at":       now.Add(time.Minute),
	})
	conn.Table("article_category").Create(map[string]any{
		"id":         3,
		"category":   "General",
		"desc":       "General desc",
		"icon":       "hash",
		"color":      "#123456",
		"slug":       "general",
		"sort":       1,
		"created_at": now,
		"updated_at": now,
	})
	conn.Table("article_category_rs").Create(map[string]any{
		"id":                  4,
		"article_id":          10,
		"article_category_id": 3,
		"effective":           1,
		"created_at":          now,
		"updated_at":          now,
	})
	conn.Table("article_user_action").Create(map[string]any{
		"id":            5,
		"user_id":       1,
		"article_id":    10,
		"liked_at":      now,
		"bookmarked_at": now,
		"watched_at":    now,
		"created_at":    now,
		"updated_at":    now,
	})
	conn.Table("articles_user_stat").Create(map[string]any{
		"id":            6,
		"article_id":    10,
		"user_id":       2,
		"reply_count":   1,
		"last_reply_at": now.Add(time.Minute),
		"created_at":    now,
		"updated_at":    now,
	})
	conn.Table("event_notification").Create(map[string]any{
		"user_id":    1,
		"event_type": eventNotification.EventTypeComment,
		"payload":    `{"articleId":10,"commentId":99,"customKey":"keep-me"}`,
		"is_read":    false,
		"created_at": now,
		"updated_at": now,
	})
	conn.Create(&reports.Entity{TargetType: reports.TargetReply, TargetId: 99, ArticleId: 10, ReporterId: 1, Status: reports.StatusOpen})

	result := BackfillTopicPostModelWithDB(conn)
	if result.Failed != 0 {
		t.Fatalf("BackfillTopicPostModelWithDB() failed=%d last=%s", result.Failed, result.LastFailed)
	}
	if result.Topics != 1 || result.Posts != 2 || result.Categories != 1 || result.TopicCategoryIndexes != 1 {
		t.Fatalf("unexpected result counts: %#v", result)
	}

	var topic topics.Entity
	if err := conn.First(&topic, 10).Error; err != nil {
		t.Fatalf("load topic: %v", err)
	}
	if topic.Title != "hello topic" || topic.PostCount != 2 || topic.ReplyCount != 1 || topic.FirstPostId == 0 || topic.LastPostId == 0 {
		t.Fatalf("topic not migrated correctly: %#v", topic)
	}

	var firstPost posts.Entity
	if err := conn.Where("topic_id = ? AND post_no = ?", 10, 1).First(&firstPost).Error; err != nil {
		t.Fatalf("load first post: %v", err)
	}
	if firstPost.Content != "first post" || firstPost.UserId != 1 {
		t.Fatalf("first post not migrated correctly: %#v", firstPost)
	}

	var replyPost posts.Entity
	if err := conn.Where("topic_id = ? AND post_no = ?", 10, 2).First(&replyPost).Error; err != nil {
		t.Fatalf("load reply post: %v", err)
	}
	if replyPost.Content != "reply body" || replyPost.UserId != 2 {
		t.Fatalf("reply post not migrated correctly: %#v", replyPost)
	}

	var mapped migrationMapping.Entity
	if err := conn.Where("scope = ? AND source_type = ? AND source_id = ?", TopicPostMigrationScope, "reply", 99).First(&mapped).Error; err != nil {
		t.Fatalf("load reply mapping: %v", err)
	}
	if mapped.TargetType != "post" || mapped.TargetId != replyPost.Id {
		t.Fatalf("reply mapping mismatch: %#v replyPost=%d", mapped, replyPost.Id)
	}

	var action topicUserAction.Entity
	if err := conn.Where("topic_id = ? AND user_id = ?", 10, 1).First(&action).Error; err != nil {
		t.Fatalf("load topic user action: %v", err)
	}
	if action.LikedAt == nil || action.BookmarkedAt == nil || action.WatchedAt == nil {
		t.Fatalf("topic user action missing timestamps: %#v", action)
	}

	var stat topicUserStat.Entity
	if err := conn.Where("topic_id = ? AND user_id = ?", 10, 2).First(&stat).Error; err != nil {
		t.Fatalf("load topic user stat: %v", err)
	}
	if stat.ReplyCount != 1 {
		t.Fatalf("topic user stat reply count=%d, want 1", stat.ReplyCount)
	}

	var payloadRaw string
	if err := conn.Table("event_notification").Select("payload").Where("id = ?", 1).Scan(&payloadRaw).Error; err != nil {
		t.Fatalf("load notification payload: %v", err)
	}
	var payloadMap map[string]any
	if err := json.Unmarshal([]byte(payloadRaw), &payloadMap); err != nil {
		t.Fatalf("unmarshal notification payload: %v", err)
	}
	if _, ok := payloadMap["articleId"]; ok {
		t.Fatalf("legacy articleId should be removed from notification payload: %#v", payloadMap)
	}
	if _, ok := payloadMap["commentId"]; ok {
		t.Fatalf("legacy commentId should be removed from notification payload: %#v", payloadMap)
	}
	if payloadMap["topicId"].(float64) != 10 || uint64(payloadMap["postId"].(float64)) != replyPost.Id {
		t.Fatalf("clean notification ids were not added: %#v replyPost=%d", payloadMap, replyPost.Id)
	}
	if payloadMap["customKey"] != "keep-me" {
		t.Fatalf("custom notification payload key was not preserved: %#v", payloadMap)
	}

	var report reports.Entity
	if err := conn.First(&report).Error; err != nil {
		t.Fatalf("load report: %v", err)
	}
	if report.TargetType != reports.TargetPost || report.TargetId != replyPost.Id || report.TopicId != 10 {
		t.Fatalf("report was not migrated to post/topic: %#v replyPost=%d", report, replyPost.Id)
	}
}

func createLegacyTopicPostTables(t *testing.T, conn *gorm.DB) {
	t.Helper()
	statements := []string{
		`CREATE TABLE articles (
			id integer primary key,
			title text,
			content text,
			description text,
			first_image_url text,
			rendered_html text,
			rendered_version integer,
			category_id text,
			user_id integer,
			posters text,
			article_status integer,
			process_status integer,
			like_count integer,
			view_count integer,
			reply_count integer,
			reply_seq integer,
			pin_weight integer,
			created_at datetime,
			updated_at datetime,
			deleted_at datetime
		)`,
		`CREATE TABLE reply (
			id integer primary key,
			article_id integer,
			reply_no integer,
			user_id integer,
			target_id integer,
			content text,
			rendered_html text,
			rendered_version integer,
			process_status integer,
			reply_id integer,
			created_at datetime,
			updated_at datetime,
			deleted_at datetime
		)`,
		`CREATE TABLE article_category (
			id integer primary key,
			category text,
			desc text,
			icon text,
			color text,
			slug text,
			sort integer,
			created_at datetime,
			updated_at datetime
		)`,
		`CREATE TABLE article_category_rs (
			id integer primary key,
			article_id integer,
			article_category_id integer,
			effective integer,
			created_at datetime,
			updated_at datetime
		)`,
		`CREATE TABLE article_user_action (
			id integer primary key,
			user_id integer,
			article_id integer,
			liked_at datetime,
			bookmarked_at datetime,
			watched_at datetime,
			created_at datetime,
			updated_at datetime
		)`,
		`CREATE TABLE articles_user_stat (
			id integer primary key,
			article_id integer,
			user_id integer,
			reply_count integer,
			last_reply_at datetime,
			created_at datetime,
			updated_at datetime
		)`,
	}
	for _, statement := range statements {
		if err := conn.Exec(statement).Error; err != nil {
			t.Fatalf("exec legacy schema %q: %v", statement, err)
		}
	}
}
