package forum

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topicUserAction"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
)

func TestTopicMetaJSONLDIncludesForumRequiredFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	c.Request = httptest.NewRequest(http.MethodGet, "https://example.com/p/post/1", nil)

	meta := buildTopicMeta(c, TopicDetailPayload{
		ID:          1,
		Title:       "讨论标题",
		Description: "讨论描述",
		URL:         "/p/post/1",
		Author:      TopicAuthorPayload{ID: 12, Username: "author"},
		CreatedAt:   time.Now().Format(time.DateTime),
		UpdatedAt:   time.Now().Format(time.DateTime),
	})

	jsonLD, ok := meta.JSONLD.(vo.ArticleJSONLD)
	if !ok {
		t.Fatalf("expected ArticleJSONLD, got %T", meta.JSONLD)
	}
	if jsonLD.Text != "讨论描述" {
		t.Fatalf("expected text field to use precomputed topic description, got %q", jsonLD.Text)
	}
	if jsonLD.Type != "DiscussionForumPosting" {
		t.Fatalf("expected DiscussionForumPosting, got %q", jsonLD.Type)
	}
	if jsonLD.Author.Name == "" {
		t.Fatal("expected author name")
	}
}

func TestTopicMetaJSONLDIncludesImageForImageOnlyTopic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	c.Request = httptest.NewRequest(http.MethodGet, "https://example.com/p/post/440", nil)

	meta := buildTopicMeta(c, TopicDetailPayload{
		ID:            440,
		Title:         "叮叮叮～又得到一个徽章",
		Description:   "",
		URL:           "/p/post/440",
		FirstImageURL: "/file/badges/earned.webp",
		Author:        TopicAuthorPayload{ID: 1, Username: "abandon1a2b"},
		CreatedAt:     time.Now().Format(time.DateTime),
		UpdatedAt:     time.Now().Format(time.DateTime),
	})

	jsonLD, ok := meta.JSONLD.(vo.ArticleJSONLD)
	if !ok {
		t.Fatalf("expected ArticleJSONLD, got %T", meta.JSONLD)
	}
	expectedText := "阅读 叮叮叮～又得到一个徽章，参与 GooseForum 的社区讨论。"
	if jsonLD.Text != expectedText {
		t.Fatalf("expected image-only topic text fallback, got %q", jsonLD.Text)
	}
	if len(jsonLD.Image) != 1 || jsonLD.Image[0] != "http://localhost/file/badges/earned.webp" {
		t.Fatalf("expected absolute inline image URL, got %#v", jsonLD.Image)
	}
	if meta.OpenGraph == nil || meta.OpenGraph.Image != "http://localhost/file/badges/earned.webp" {
		t.Fatalf("expected OpenGraph image to use first inline image, got %#v", meta.OpenGraph)
	}
	if meta.Twitter == nil || meta.Twitter.Image != "http://localhost/file/badges/earned.webp" {
		t.Fatalf("expected Twitter image to use first inline image, got %#v", meta.Twitter)
	}
}

func TestDraftTopicCanOnlyBeViewedByAuthor(t *testing.T) {
	draft := &topics.Entity{Id: 1, UserId: 10, Status: 0, ProcessStatus: 0}

	if !canViewTopic(draft, 10) {
		t.Fatal("expected draft author to view draft topic")
	}
	if canViewTopic(draft, 11) {
		t.Fatal("expected other users to be blocked from draft topic")
	}
	if canViewTopic(draft, 0) {
		t.Fatal("expected guests to be blocked from draft topic")
	}
}

func TestDraftTopicViewIsNotCounted(t *testing.T) {
	draft := &topics.Entity{Id: 1, UserId: 10, Status: 0, ProcessStatus: 0}
	published := &topics.Entity{Id: 2, UserId: 10, Status: 1, ProcessStatus: 0}
	blocked := &topics.Entity{Id: 3, UserId: 10, Status: 1, ProcessStatus: 1}

	if shouldCountTopicView(draft) {
		t.Fatal("expected draft topic views to be ignored")
	}
	if !shouldCountTopicView(published) {
		t.Fatal("expected published normal topic views to be counted")
	}
	if shouldCountTopicView(blocked) {
		t.Fatal("expected blocked topic views to be ignored")
	}
}

func TestBuildTopicDetailPropsReadsTopicPostTables(t *testing.T) {
	preferences.Set("db.default.connection", "sqlite")
	preferences.Set("db.default.path", ":memory:")
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&topics.Entity{}, &posts.Entity{}, &category.Entity{}, &users.EntityComplete{}, &topicUserAction.Entity{}); err != nil {
		t.Fatalf("migrate topic detail tables: %v", err)
	}
	hotdataserve.ClearTopicCategoryCache()

	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	topicID := uint64(990010)
	firstPostID := uint64(990100)
	replyPostID := uint64(990101)
	authorID := uint64(990001)
	replyerID := uint64(990002)
	categoryID := uint64(990003)
	t.Cleanup(func() {
		conn.Unscoped().Where("topic_id = ?", topicID).Delete(&posts.Entity{})
		conn.Unscoped().Delete(&topics.Entity{}, topicID)
		conn.Unscoped().Delete(&category.Entity{}, categoryID)
		conn.Unscoped().Delete(&users.EntityComplete{}, []uint64{authorID, replyerID})
		conn.Where("user_id = ? AND topic_id = ?", authorID, topicID).Delete(&topicUserAction.Entity{})
		hotdataserve.ClearTopicCategoryCache()
	})
	conn.Unscoped().Where("topic_id = ?", topicID).Delete(&posts.Entity{})
	conn.Unscoped().Delete(&topics.Entity{}, topicID)
	conn.Unscoped().Delete(&category.Entity{}, categoryID)
	conn.Unscoped().Delete(&users.EntityComplete{}, []uint64{authorID, replyerID})
	conn.Where("user_id = ? AND topic_id = ?", authorID, topicID).Delete(&topicUserAction.Entity{})
	conn.Create(&users.EntityComplete{Id: authorID, Username: "author"})
	conn.Create(&users.EntityComplete{Id: replyerID, Username: "replyer"})
	conn.Create(&category.Entity{Id: categoryID, Name: "General", Slug: "general"})
	topic := topics.Entity{
		Id:            topicID,
		Title:         "topic title",
		Excerpt:       "topic excerpt",
		CategoryIds:   []uint64{categoryID},
		UserId:        authorID,
		Status:        1,
		ProcessStatus: 0,
		ReplyCount:    1,
		PostSeq:       2,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	conn.Create(&topic)
	firstPost := posts.Entity{Id: firstPostID, TopicId: topicID, PostNo: 1, UserId: authorID, Content: "first", RenderedHTML: "<p>first</p>", RenderedVersion: 1, CreatedAt: now, UpdatedAt: now}
	conn.Create(&firstPost)
	conn.Model(&topics.Entity{}).Where("id = ?", topicID).Update("first_post_id", firstPost.Id)
	conn.Create(&posts.Entity{Id: replyPostID, TopicId: topicID, PostNo: 2, UserId: replyerID, Content: "reply", RenderedHTML: "<p>reply</p>", RenderedVersion: 1, CreatedAt: now.Add(time.Minute), UpdatedAt: now.Add(time.Minute)})
	conn.Create(&topicUserAction.Entity{UserId: authorID, TopicId: topicID, LikedAt: &now, WatchedAt: &now})

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	c.Request = httptest.NewRequest(http.MethodGet, "/p/post/990010", nil)
	props := buildTopicDetailProps(c, &topic, &firstPost)

	if props.Topic.ID != topicID || props.Topic.MaxPostNo != 2 {
		t.Fatalf("topic payload mismatch: %#v", props.Topic)
	}
	if len(props.Topic.Categories) != 1 || props.Topic.Categories[0].Name != "General" {
		t.Fatalf("categories mismatch: %#v", props.Topic.Categories)
	}
	if len(props.PostStream.Posts) != 2 {
		t.Fatalf("post stream length = %d, want 2", len(props.PostStream.Posts))
	}
	if props.PostStream.Posts[0].ID != firstPostID || props.PostStream.Posts[0].PostNo != 1 || props.PostStream.Posts[0].RenderedContent != "<p>first</p>" {
		t.Fatalf("first post payload mismatch: %#v", props.PostStream.Posts[0])
	}
	if props.PostStream.Posts[1].ID != replyPostID || props.PostStream.Posts[1].PostNo != 2 {
		t.Fatalf("reply post payload mismatch: %#v", props.PostStream.Posts[1])
	}
	if props.PostStream.BeforePostNo != 1 || props.PostStream.AfterPostNo != 2 || props.PostStream.MaxPostNo != 2 {
		t.Fatalf("post stream cursor mismatch: %#v", props.PostStream)
	}
}

func TestPostWindowDefaultLoadsFromFirstPost(t *testing.T) {
	preferences.Set("db.default.connection", "sqlite")
	preferences.Set("db.default.path", ":memory:")
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&topics.Entity{}, &posts.Entity{}, &users.EntityComplete{}); err != nil {
		t.Fatalf("migrate reply window tables: %v", err)
	}

	now := time.Date(2026, 7, 7, 13, 0, 0, 0, time.UTC)
	topicID := uint64(991010)
	firstPostID := uint64(991100)
	replyPostID := uint64(991101)
	userID := uint64(991001)
	t.Cleanup(func() {
		conn.Unscoped().Where("topic_id = ?", topicID).Delete(&posts.Entity{})
		conn.Unscoped().Delete(&topics.Entity{}, topicID)
		conn.Unscoped().Delete(&users.EntityComplete{}, userID)
	})
	conn.Unscoped().Where("topic_id = ?", topicID).Delete(&posts.Entity{})
	conn.Unscoped().Delete(&topics.Entity{}, topicID)
	conn.Unscoped().Delete(&users.EntityComplete{}, userID)
	conn.Create(&users.EntityComplete{Id: userID, Username: "author"})
	conn.Create(&topics.Entity{
		Id:            topicID,
		Title:         "topic",
		UserId:        userID,
		Status:        1,
		ProcessStatus: 0,
		ReplyCount:    1,
		PostSeq:       2,
		FirstPostId:   firstPostID,
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	conn.Create(&posts.Entity{Id: firstPostID, TopicId: topicID, PostNo: 1, UserId: userID, Content: "first", CreatedAt: now, UpdatedAt: now})
	conn.Create(&posts.Entity{Id: replyPostID, TopicId: topicID, PostNo: 2, UserId: userID, Content: "reply", CreatedAt: now.Add(time.Minute), UpdatedAt: now.Add(time.Minute)})

	res := PostWindow(component.BetterRequest[PostWindowReq]{
		Params: PostWindowReq{
			TopicID: topicID,
			Limit:   1,
		},
	})
	payload, ok := res.Data.Result.(PostWindowPayload)
	if !ok {
		t.Fatalf("result type = %T, want PostWindowPayload", res.Data.Result)
	}
	if len(payload.Posts) != 1 || payload.Posts[0].ID != firstPostID || payload.Posts[0].PostNo != 1 {
		t.Fatalf("posts = %#v, want default first post window", payload.Posts)
	}
	if payload.BeforePostNo != 1 || payload.AfterPostNo != 1 {
		t.Fatalf("cursor payload = %#v", payload)
	}
	if payload.Total != 2 || payload.MaxPostNo != 2 {
		t.Fatalf("stream totals = %#v", payload)
	}
}

func TestPostWindowAnchorPostNoCanLoadFirstPost(t *testing.T) {
	preferences.Set("db.default.connection", "sqlite")
	preferences.Set("db.default.path", ":memory:")
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&topics.Entity{}, &posts.Entity{}, &users.EntityComplete{}); err != nil {
		t.Fatalf("migrate post window tables: %v", err)
	}

	now := time.Date(2026, 7, 7, 14, 0, 0, 0, time.UTC)
	topicID := uint64(993010)
	firstPostID := uint64(993100)
	replyPostID := uint64(993101)
	userID := uint64(993001)
	t.Cleanup(func() {
		conn.Unscoped().Where("topic_id = ?", topicID).Delete(&posts.Entity{})
		conn.Unscoped().Delete(&topics.Entity{}, topicID)
		conn.Unscoped().Delete(&users.EntityComplete{}, userID)
	})
	conn.Unscoped().Where("topic_id = ?", topicID).Delete(&posts.Entity{})
	conn.Unscoped().Delete(&topics.Entity{}, topicID)
	conn.Unscoped().Delete(&users.EntityComplete{}, userID)
	conn.Create(&users.EntityComplete{Id: userID, Username: "author"})
	conn.Create(&topics.Entity{
		Id:          topicID,
		Title:       "topic",
		UserId:      userID,
		Status:      1,
		ReplyCount:  1,
		PostSeq:     2,
		FirstPostId: firstPostID,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	conn.Create(&posts.Entity{Id: firstPostID, TopicId: topicID, PostNo: 1, UserId: userID, Content: "first", CreatedAt: now, UpdatedAt: now})
	conn.Create(&posts.Entity{Id: replyPostID, TopicId: topicID, PostNo: 2, UserId: userID, Content: "reply", CreatedAt: now.Add(time.Minute), UpdatedAt: now.Add(time.Minute)})

	res := PostWindow(component.BetterRequest[PostWindowReq]{
		Params: PostWindowReq{
			TopicID:      topicID,
			AnchorPostNo: 1,
			Limit:        20,
		},
	})
	payload, ok := res.Data.Result.(PostWindowPayload)
	if !ok {
		t.Fatalf("result type = %T, want PostWindowPayload", res.Data.Result)
	}
	if len(payload.Posts) != 2 || payload.Posts[0].PostNo != 1 || payload.Posts[1].PostNo != 2 {
		t.Fatalf("payload posts = %#v, want first post and reply", payload.Posts)
	}
	if payload.HasBefore {
		t.Fatalf("first post anchor should not have before window: %#v", payload)
	}
}

func TestPostWindowAnchorPostNoFallsForwardAcrossDeletedReplies(t *testing.T) {
	preferences.Set("db.default.connection", "sqlite")
	preferences.Set("db.default.path", ":memory:")
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&topics.Entity{}, &posts.Entity{}, &users.EntityComplete{}); err != nil {
		t.Fatalf("migrate reply window tables: %v", err)
	}

	now := time.Date(2026, 7, 7, 13, 0, 0, 0, time.UTC)
	topicID := uint64(992010)
	firstPostID := uint64(992100)
	userID := uint64(992001)
	t.Cleanup(func() {
		conn.Unscoped().Where("topic_id = ?", topicID).Delete(&posts.Entity{})
		conn.Unscoped().Delete(&topics.Entity{}, topicID)
		conn.Unscoped().Delete(&users.EntityComplete{}, userID)
	})
	conn.Unscoped().Where("topic_id = ?", topicID).Delete(&posts.Entity{})
	conn.Unscoped().Delete(&topics.Entity{}, topicID)
	conn.Unscoped().Delete(&users.EntityComplete{}, userID)
	conn.Create(&users.EntityComplete{Id: userID, Username: "author"})
	conn.Create(&topics.Entity{
		Id:            topicID,
		Title:         "topic",
		UserId:        userID,
		Status:        1,
		ProcessStatus: 0,
		ReplyCount:    1,
		PostSeq:       4,
		FirstPostId:   firstPostID,
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	conn.Create(&posts.Entity{Id: firstPostID, TopicId: topicID, PostNo: 1, UserId: userID, Content: "first", CreatedAt: now, UpdatedAt: now})
	conn.Create(&posts.Entity{Id: 992104, TopicId: topicID, PostNo: 4, UserId: userID, Content: "reply after deleted replies", CreatedAt: now.Add(time.Minute), UpdatedAt: now.Add(time.Minute)})

	res := PostWindow(component.BetterRequest[PostWindowReq]{
		Params: PostWindowReq{
			TopicID:      topicID,
			AnchorPostNo: 2,
			Limit:        20,
		},
	})
	payload, ok := res.Data.Result.(PostWindowPayload)
	if !ok {
		t.Fatalf("result type = %T, want PostWindowPayload", res.Data.Result)
	}
	if len(payload.Posts) != 2 || payload.Posts[0].PostNo != 1 || payload.Posts[1].PostNo != 4 || payload.BeforePostNo != 1 || payload.AfterPostNo != 4 || payload.MaxPostNo != 4 {
		t.Fatalf("payload = %#v, want first post plus nearest remaining post no 4", payload)
	}
}
