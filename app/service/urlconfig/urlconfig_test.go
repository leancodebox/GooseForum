package urlconfig

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/preferences"
)

func TestDefaultAvatarUsesCDNWhenConfigured(t *testing.T) {
	old := preferences.GetString("app.cdn_url", "")
	t.Cleanup(func() {
		preferences.Set("app.cdn_url", old)
	})

	preferences.Set("app.cdn_url", "")
	if got := GetDefaultAvatar(); got != "/static/pic/default-avatar.webp" {
		t.Fatalf("default avatar = %q, want local path", got)
	}

	preferences.Set("app.cdn_url", "https://cdn.example.com")
	if got := GetDefaultAvatar(); got != "https://cdn.example.com/static/pic/default-avatar.webp" {
		t.Fatalf("cdn default avatar = %q, want CDN path", got)
	}
}

func TestFilePath(t *testing.T) {
	if got := FilePath("avatar.webp"); got != "/file/img/avatar.webp" {
		t.Fatalf("FilePath = %q, want /file/img/avatar.webp", got)
	}
	if got := FilePath("/nested/avatar.webp"); got != "/file/img/nested/avatar.webp" {
		t.Fatalf("FilePath nested = %q, want normalized image path", got)
	}
}

func TestStaticRoutes(t *testing.T) {
	tests := map[string]string{
		"home":          Home(),
		"post":          Post(),
		"docs":          Docs(),
		"links":         Links(),
		"sponsors":      Sponsors(),
		"publish":       Publish(),
		"search":        Search(),
		"register":      Register(),
		"login":         Login(),
		"messages":      Messages(),
		"drafts":        Drafts(),
		"settings":      Settings(),
		"notifications": Notifications(),
		"activate":      Activate(),
		"resetPassword": ResetPassword(),
		"admin":         Admin(),
		"rss":           Rss(),
	}

	want := map[string]string{
		"home":          "/",
		"post":          "/p/post",
		"docs":          "/docs",
		"links":         "/links",
		"sponsors":      "/sponsors",
		"publish":       "/publish",
		"search":        "/search",
		"register":      "/login",
		"login":         "/login",
		"messages":      "/messages",
		"drafts":        "/drafts",
		"settings":      "/settings",
		"notifications": "/notifications",
		"activate":      "/activate",
		"resetPassword": "/reset-password",
		"admin":         "/admin",
		"rss":           "/rss.xml",
	}

	for name, got := range tests {
		if got != want[name] {
			t.Fatalf("%s route = %q, want %q", name, got, want[name])
		}
	}
}

func TestDynamicRoutes(t *testing.T) {
	if got := PostDetail(42); got != "/p/post/42" {
		t.Fatalf("PostDetail = %q, want /p/post/42", got)
	}
	if got := User("alice"); got != "/u/alice" {
		t.Fatalf("User = %q, want /u/alice", got)
	}
	if got := DocsProject("goose"); got != "/docs/goose" {
		t.Fatalf("DocsProject = %q, want /docs/goose", got)
	}
	if got := DocsContent("goose", "v1", "intro"); got != "/docs/goose/v1/intro" {
		t.Fatalf("DocsContent = %q, want /docs/goose/v1/intro", got)
	}
}
