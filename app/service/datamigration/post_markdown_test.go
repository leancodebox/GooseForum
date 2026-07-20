package datamigration

import (
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"gorm.io/gorm"
)

func TestRebuildPostMarkdownMigratesHistoricalPosts(t *testing.T) {
	conn, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := conn.AutoMigrate(&posts.Entity{}); err != nil {
		t.Fatalf("migrate posts: %v", err)
	}

	old := posts.Entity{Id: 1, TopicId: 1, PostNo: 1, Content: "[external](https://example.com)", RenderedHTML: "<p>old</p>", RenderedVersion: 3}
	current := posts.Entity{Id: 2, TopicId: 1, PostNo: 2, Content: "current", RenderedHTML: "<p>keep</p>", RenderedVersion: markdown2html.GetPostVersion()}
	if err := conn.Create(&[]posts.Entity{old, current}).Error; err != nil {
		t.Fatalf("create posts: %v", err)
	}

	result := RebuildPostMarkdownWithDB(conn)
	if result.Processed != 1 || result.Failed != 0 {
		t.Fatalf("migration result = %#v", result)
	}

	var got []posts.Entity
	if err := conn.Order("id").Find(&got).Error; err != nil {
		t.Fatalf("load posts: %v", err)
	}
	if got[0].RenderedVersion != markdown2html.GetPostVersion() || !strings.Contains(got[0].RenderedHTML, `rel="nofollow ugc noopener noreferrer"`) {
		t.Fatalf("historical post not rebuilt: %#v", got[0])
	}
	if got[1].RenderedHTML != "<p>keep</p>" {
		t.Fatalf("current post was rebuilt: %#v", got[1])
	}
}
