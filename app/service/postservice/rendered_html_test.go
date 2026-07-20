package postservice

import (
	"strings"
	"testing"

	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
)

func TestEnsureRenderedHTMLRebuildsAndSavesStalePost(t *testing.T) {
	post := &posts.Entity{
		Id:              1,
		Content:         "[external](https://example.com)",
		RenderedHTML:    "<p>stale</p>",
		RenderedVersion: markdown2html.GetPostVersion() - 1,
	}
	saved := false

	got, err := ensureRenderedHTML(post, func(entity *posts.Entity) error {
		saved = true
		return nil
	})

	if err != nil || !saved {
		t.Fatalf("ensureRenderedHTML() err = %v, saved = %v", err, saved)
	}
	if strings.Contains(got, "stale") || !strings.Contains(got, `rel="nofollow ugc noopener noreferrer"`) {
		t.Fatalf("ensureRenderedHTML() = %q, want current post rendering", got)
	}
}

func TestEnsureRenderedHTMLReusesCurrentPostWithoutSaving(t *testing.T) {
	post := &posts.Entity{Id: 1, RenderedHTML: "<p>current</p>", RenderedVersion: markdown2html.GetPostVersion()}

	got, err := ensureRenderedHTML(post, func(entity *posts.Entity) error {
		t.Fatal("current post should not be saved")
		return nil
	})

	if err != nil || got != post.RenderedHTML {
		t.Fatalf("ensureRenderedHTML() = %q, %v", got, err)
	}
}
