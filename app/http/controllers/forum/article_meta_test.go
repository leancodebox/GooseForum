package forum

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
)

func TestArticleMetaJSONLDIncludesForumRequiredFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	c.Request = httptest.NewRequest("GET", "https://example.com/p/post/1", nil)

	meta := buildArticleMeta(c, ArticlePayload{
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
	if jsonLD.Text != "正文内容 重点" {
		t.Fatalf("expected text field to contain plain article body, got %q", jsonLD.Text)
	}
	if jsonLD.Type != "DiscussionForumPosting" {
		t.Fatalf("expected DiscussionForumPosting, got %q", jsonLD.Type)
	}
	if jsonLD.Author.Name == "" {
		t.Fatal("expected author name")
	}
}

func TestDraftArticleCanOnlyBeViewedByAuthor(t *testing.T) {
	draft := &articles.Entity{Id: 1, UserId: 10, ArticleStatus: 0, ProcessStatus: 0}

	if !canViewArticle(draft, 10, false) {
		t.Fatal("expected draft author to view draft article")
	}
	if canViewArticle(draft, 11, false) {
		t.Fatal("expected other users to be blocked from draft article")
	}
	if canViewArticle(draft, 0, false) {
		t.Fatal("expected guests to be blocked from draft article")
	}
}

func TestDraftArticleViewIsNotCounted(t *testing.T) {
	draft := &articles.Entity{Id: 1, UserId: 10, ArticleStatus: 0, ProcessStatus: 0}
	published := &articles.Entity{Id: 2, UserId: 10, ArticleStatus: 1, ProcessStatus: 0}
	blocked := &articles.Entity{Id: 3, UserId: 10, ArticleStatus: 1, ProcessStatus: 1}

	if shouldCountArticleView(draft) {
		t.Fatal("expected draft article views to be ignored")
	}
	if !shouldCountArticleView(published) {
		t.Fatal("expected published normal article views to be counted")
	}
	if shouldCountArticleView(blocked) {
		t.Fatal("expected blocked article views to be ignored")
	}
}
