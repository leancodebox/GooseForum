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

func TestArticleMetaJSONLDIncludesForumRequiredFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	c.Request = httptest.NewRequest(http.MethodGet, "https://example.com/p/post/1", nil)

	meta := buildArticleMeta(c, TopicDetailPayload{
		ID:          1,
		Title:       "讨论标题",
		Description: "讨论描述",
		URL:         "/p/post/1",
		HTML:        "<p>正文内容 <strong>重点</strong></p>",
		Author:      TopicAuthorPayload{ID: 12, Username: "author"},
		CreatedAt:   time.Now().Format(time.DateTime),
		UpdatedAt:   time.Now().Format(time.DateTime),
	})

	jsonLD, ok := meta.JSONLD.(vo.ArticleJSONLD)
	if !ok {
		t.Fatalf("expected ArticleJSONLD, got %T", meta.JSONLD)
	}
	if jsonLD.Text != "讨论描述" {
		t.Fatalf("expected text field to use precomputed article description, got %q", jsonLD.Text)
	}
	if jsonLD.Type != "DiscussionForumPosting" {
		t.Fatalf("expected DiscussionForumPosting, got %q", jsonLD.Type)
	}
	if jsonLD.Author.Name == "" {
		t.Fatal("expected author name")
	}
}

func TestArticleMetaJSONLDIncludesImageForImageOnlyArticle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	c.Request = httptest.NewRequest(http.MethodGet, "https://example.com/p/post/440", nil)

	meta := buildArticleMeta(c, TopicDetailPayload{
		ID:            440,
		Title:         "叮叮叮～又得到一个徽章",
		Description:   "",
		URL:           "/p/post/440",
		FirstImageURL: "/file/badges/earned.webp",
		HTML:          `<p><img src="/file/badges/fallback.webp" alt="徽章"></p>`,
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
		t.Fatalf("expected image-only article text fallback, got %q", jsonLD.Text)
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
	hotdataserve.ClearArticleCategoryCache()

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
		hotdataserve.ClearArticleCategoryCache()
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

	if props.Topic.ID != topicID || props.Topic.HTML != "<p>first</p>" || props.Topic.MaxPostNo != 1 {
		t.Fatalf("article payload mismatch: %#v", props.Topic)
	}
	if len(props.Topic.Categories) != 1 || props.Topic.Categories[0].Name != "General" {
		t.Fatalf("categories mismatch: %#v", props.Topic.Categories)
	}
	if len(props.Posts) != 1 || props.Posts[0].ID != replyPostID || props.Posts[0].PostNo != 1 {
		t.Fatalf("posts mismatch: %#v", props.Posts)
	}
}

func TestPostWindowSkipsFirstPostInCursors(t *testing.T) {
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
			Tail:    true,
			Limit:   50,
		},
	})
	payload, ok := res.Data.Result.(PostWindowPayload)
	if !ok {
		t.Fatalf("result type = %T, want PostWindowPayload", res.Data.Result)
	}
	if len(payload.Posts) != 1 || payload.Posts[0].ID != replyPostID {
		t.Fatalf("posts = %#v, want only reply post", payload.Posts)
	}
	if payload.BeforeCursor != replyPostID || payload.AfterCursor != replyPostID || payload.BeforePostNo != 1 || payload.AfterPostNo != 1 {
		t.Fatalf("cursor payload = %#v", payload)
	}
}
